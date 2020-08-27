package nacos

import (
	"fmt"
	"time"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/http_agent"
	"github.com/stretchr/testify/mock"
)

var serverConfig = constant.ServerConfig{
	ContextPath: "/nacos",
	Port:        8848,
	IpAddr:      "127.0.0.1",
}

var clientConfigTest = constant.ClientConfig{
	BeatInterval:   10000,
	TimeoutMs:      10000,
	ListenInterval: 20000,
}

func createConfigClientTest() (*config_client.ConfigClient, error) {
	nc := nacos_client.NacosClient{}
	if err := nc.SetServerConfig([]constant.ServerConfig{serverConfig}); err != nil {
		return nil, err
	}
	if err := nc.SetClientConfig(clientConfigTest); err != nil {
		return nil, err
	}
	if err := nc.SetHttpAgent(&http_agent.HttpAgent{}); err != nil {
		return nil, err
	}
	client, err := config_client.NewConfigClient(&nc)
	return &client, err
}

func Example_NacosDatasource_CustomizeClient() {
	client, err := createConfigClientTest()
	if err != nil {
		fmt.Printf("Fail to create client, err: %+v", err)
		return
	}
	h := &datasource.MockPropertyHandler{}
	h.On("isPropertyConsistent", mock.Anything).Return(true)
	h.On("Handle", mock.Anything).Return(nil)
	nds, err := NewNacosDataSource(client, "sentinel-go", "system-rules", h)
	if err != nil {
		fmt.Printf("Fail to create nacos data source client, err: %+v", err)
		return
	}
	err = nds.Initialize()
	if err != nil {
		fmt.Printf("Fail to initialize nacos data source client, err: %+v", err)
		return
	}

	time.Sleep(time.Second * 10)
	nds.Close()
	fmt.Println("Nacos datasource is Closed")
}
