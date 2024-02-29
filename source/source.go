package source

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Source struct {
	sdk.UnimplementedSource

	config                    Config
	client                    *azeventhubs.ConsumerClient
	processor                 *azeventhubs.Processor
	partitionReadErrorChannel chan error
	readBuffer                chan sdk.Record
}

func New() sdk.Source {
	return sdk.SourceWithMiddleware(&Source{
		partitionReadErrorChannel: make(chan error, 1),
		readBuffer:                make(chan sdk.Record, 1),
	}, sdk.DefaultSourceMiddleware()...)
}

func (s *Source) Parameters() map[string]sdk.Parameter {
	return Config{}.Parameters()
}

func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	sdk.Logger(ctx).Info().Msg("Configuring Source...")
	err := sdk.Util.ParseConfig(cfg, &s.config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	return nil
}

func (s *Source) Open(ctx context.Context, pos sdk.Position) error {
	defaultAzureCred, err := azidentity.NewClientSecretCredential(s.config.AzureTenantID, s.config.AzureClientID, s.config.AzureClientSecret, nil)
	if err != nil {
		return err
	}

	s.client, err = azeventhubs.NewConsumerClient(s.config.EventHubNameSpace, s.config.EventHubName, azeventhubs.DefaultConsumerGroup, defaultAzureCred, nil)
	if err != nil {
		return err
	}

	s.processor, err = azeventhubs.NewProcessor(s.client, newCheckpointStore(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	// dispatch the processors, when they read in a new record, push it to our read buffer
	go s.dispatchPartitionClients(ctx)

	select {
	case err := <-s.partitionReadErrorChannel:
		return sdk.Record{}, err
	case rec := <-s.readBuffer:
		return rec, nil
	case <-ctx.Done():
		return sdk.Record{}, ctx.Err()
	}
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	return nil
}

func (s *Source) Teardown(ctx context.Context) error {
	if s.client != nil {
		s.client.Close(ctx)
	}
	return nil
}

func (s *Source) dispatchPartitionClients(ctx context.Context) {
	for {
		processorPartitionClient := s.processor.NextPartitionClient(ctx)

		if processorPartitionClient == nil {
			// Processor has stopped
			break
		}

		go func() {
			defer func() {
				// 3/3 [END] Do cleanup here, like shutting down database clients
				// or other resources used for processing this partition.
				processorPartitionClient.Close(ctx)
			}()

			for {
				// Wait up to a minute for 100 events, otherwise returns whatever we collected during that time.
				receiveCtx, cancelReceive := context.WithTimeout(context.TODO(), time.Minute)
				events, err := processorPartitionClient.ReceiveEvents(receiveCtx, 100, nil)
				cancelReceive()

				if err != nil && !errors.Is(err, context.DeadlineExceeded) {
					var eventHubError *azeventhubs.Error

					if errors.As(err, &eventHubError) && eventHubError.Code == azeventhubs.ErrorCodeOwnershipLost {
						return
					}

					s.partitionReadErrorChannel <- err
					return
				}

				if len(events) == 0 {
					continue
				}

				for _, event := range events {
					// create record per event
					rec := sdk.Util.Source.NewRecordCreate(
						nil,
						nil,
						sdk.RawData(*event.MessageID),
						sdk.RawData(event.Body))

					s.readBuffer <- rec
				}

				// Updates the checkpoint with the latest event received. If processing needs to restart
				// it will restart from this point, automatically.
				if err := processorPartitionClient.UpdateCheckpoint(ctx, events[len(events)-1], nil); err != nil {
					s.partitionReadErrorChannel <- err
					return
				}
			}
		}()
	}
}
