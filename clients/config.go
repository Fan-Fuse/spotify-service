package clients

import (
	"context"
	"time"

	"github.com/Fan-Fuse/config-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var configClient proto.ConfigServiceClient

type configEntry struct {
	Key   string
	Value string
}

// We are defining the keys this service wants here
var Config = []configEntry{
	{Key: "APP_ENV", Value: ""},
	{Key: "APP_VERSION", Value: ""},
	{Key: "USER_REGISTRATION_OPEN", Value: ""},
}

// NewConfigServiceClient creates a new ConfigServiceClient.
func InitConfig(addr string) {
	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	configClient = proto.NewConfigServiceClient(cc)

	// Wait for the connection to be established
	waitForConfigService()

	// Get the initial values for the keys
	getKeys()

	// Subscribe to the keys
	go subscribeToKeys()
}

func waitForConfigService() {
	zap.S().Info("Waiting for config service...")
	for {
		_, err := configClient.GetKey(context.Background(), &proto.GetKeyRequest{Key: "test"})
		if err == nil {
			return
		} else {
			time.Sleep(time.Second * 5)
		}
	}
}

// getKeys gets the initial values for the keys
func getKeys() {
	for i := range Config {
		resp, err := configClient.GetKey(context.Background(), &proto.GetKeyRequest{Key: Config[i].Key})
		if err != nil {
			zap.S().Fatal("Error getting key", zap.String("key", Config[i].Key))
		}
		Config[i].Value = resp.Value // Modify the actual element in the Config slice
	}
}

func GetKey(key string) string {
	for i := range Config {
		if Config[i].Key == key {
			return Config[i].Value
		}
	}
	return ""
}

// subscribeToKeys subscribes to the keys in a background goroutine, updating the Config slice
func subscribeToKeys() {
	stream, err := configClient.Subscribe(context.Background(), &proto.SubscribeRequest{
		Keys: []string{"APP_ENV", "APP_VERSION", "USER_REGISTRATION_OPEN"},
	})
	if err != nil {
		zap.S().Fatal("Error subscribing to keys")
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			zap.S().Fatal("Error receiving key update")
		}

		for i := range Config {
			// TODO: React to the Key change, eventually fully reloading the service
			if Config[i].Key == resp.Key {
				Config[i].Value = resp.Value // Modify the actual element in the Config slice
				zap.S().Info(zap.String("event", "KEY_UPDATE"), zap.String("value", resp.Key))
			}
		}
	}
}
