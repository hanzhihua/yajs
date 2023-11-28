package aliyun

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	ecs "github.com/alibabacloud-go/ecs-20140526/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/hanzhihua/yajs/config"
	"github.com/hanzhihua/yajs/utils"
	"os"
)

func getAllServer() ([]*config.Server, error) {

	openapiConfig := &openapi.Config{}
	openapiConfig.AccessKeyId = tea.String(os.Getenv("ACCESS_KEY_ID"))
	openapiConfig.AccessKeySecret = tea.String(os.Getenv("ACCESS_KEY_SECRET"))
	openapiConfig.Endpoint = tea.String("ecs.cn-shanghai.aliyuncs.com")
	openapiConfig.ConnectTimeout = tea.Int(5000)
	openapiConfig.ReadTimeout = tea.Int(5000)
	client, _err := ecs.NewClient(openapiConfig)
	if _err != nil {
		return nil, _err
	}
	regionId := "cn-shanghai"
	describeInstancesRequest := &ecs.DescribeInstancesRequest{
		PageSize: tea.Int32(100),
		RegionId: &regionId,
	}
	resp, _err := client.DescribeInstances(describeInstancesRequest)
	if _err != nil {
		return nil, _err
	}

	ips, err := utils.GetLocalIPs()
	if err != nil {
		return nil, err
	}
	var servers []*config.Server
	instances := resp.Body.Instances.Instance
	for _, instance := range instances {
		server := config.Server{
			Name: *instance.InstanceName,
			IP:   *instance.VpcAttributes.PrivateIpAddress.IpAddress[0],
			Port: 22,
		}
		if !utils.ContainsStr(ips, server.IP) {
			servers = append(servers, &server)
		} else {
			utils.Logger.Warningf("ignore jumpserver:%v", server)
		}

	}
	return servers, nil

}
