package source

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
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
	partitionClients          []*azeventhubs.PartitionClient
	dispatched                bool
}

func New() sdk.Source {
	return sdk.SourceWithMiddleware(&Source{
		partitionReadErrorChannel: make(chan error, 1),
		readBuffer:                make(chan sdk.Record, 10000),
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

	ehProps, err := s.client.GetEventHubProperties(ctx, nil)
	if err != nil {
		return err
	}

	for _, partitionID := range ehProps.PartitionIDs {
		partitionClient, err := s.client.NewPartitionClient(partitionID, &azeventhubs.PartitionClientOptions{
			StartPosition: azeventhubs.StartPosition{
				Earliest: to.Ptr[bool](true),
			},
		})
		if err != nil {
			return err
		}

		s.partitionClients = append(s.partitionClients, partitionClient)
	}

	// go func() {
	// 	s.dispatchPartitionClients(ctx)

	// 	err := <-s.partitionReadErrorChannel
	// 	if err != nil {
	// 		sdk.Logger(ctx).Err(err)
	// 	}
	// }()
	return nil
}

func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	if !s.dispatched {
		go s.dispatchPartitionClients(ctx)
	}

	select {
	case err := <-s.partitionReadErrorChannel:
		if err != nil {
			return sdk.Record{}, err
		}
		return sdk.Record{}, ctx.Err()
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
		err := s.client.Close(ctx)
		if err != nil {
			return err
		}
	}

	for _, client := range s.partitionClients {
		client.Close(ctx)
	}

	return nil
}

func (s *Source) dispatchPartitionClients(ctx context.Context) {
	s.dispatched = true

	for _, client := range s.partitionClients {
		client := client
		go func() {
			// Wait up to a 500ms for 100 events, otherwise returns whatever we collected during that time.
			receiveCtx, cancelReceive := context.WithTimeout(ctx, time.Second*1)
			events, err := client.ReceiveEvents(receiveCtx, 1000, nil)
			defer cancelReceive()

			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				fmt.Println(err)
				s.partitionReadErrorChannel <- err
				return
			}

			if len(events) == 0 {
				return
			}

			for _, event := range events {
				position := fmt.Sprintf("%d", event.SequenceNumber)
				if event.PartitionKey != nil {
					position = fmt.Sprintf("%s-%d", *event.PartitionKey, event.SequenceNumber)
				}

				rec := sdk.Util.Source.NewRecordCreate(
					sdk.Position(position),
					nil,
					sdk.RawData(*event.MessageID),
					sdk.RawData(event.Body))

				s.readBuffer <- rec
			}
		}()
	}

	select {
	case err := <-s.partitionReadErrorChannel:
		if err != nil {
			// log error and close out
			return
		}
	case <-ctx.Done():
		s.dispatched = false
		return
	}

}
