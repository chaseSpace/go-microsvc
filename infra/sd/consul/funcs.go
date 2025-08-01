package consul

import capi "github.com/hashicorp/consul/api"

func healthCheckAttr(uniqueName, tcpAddr string) *capi.AgentServiceCheck {
	return &capi.AgentServiceCheck{
		CheckID:  uniqueName,
		Interval: "5s",
		Timeout:  "3s",
		TCP:      tcpAddr,
		//Status:   capi.HealthPassing, // 不建议指定初始状态，而是等待consul自动检测
		DeregisterCriticalServiceAfter: "30s", // 下线超时后自动注销
	}
}
