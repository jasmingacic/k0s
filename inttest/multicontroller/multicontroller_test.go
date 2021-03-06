/*
Copyright 2021 k0s authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package multicontroller

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/k0sproject/k0s/inttest/common"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MultiControllerSuite struct {
	common.FootlooseSuite
}

func (s *MultiControllerSuite) getMainIPAddress() string {
	ssh, err := s.SSH("controller0")
	s.Require().NoError(err)
	defer ssh.Disconnect()

	ipAddress, err := ssh.ExecWithOutput("hostname -i")
	s.Require().NoError(err)
	return ipAddress
}

func (s *MultiControllerSuite) TestK0sGetsUp() {
	ipAddress := s.getMainIPAddress()
	s.T().Logf("ip address: %s", ipAddress)

	s.putFile("controller0", "/tmp/k0s.yaml", fmt.Sprintf(k0sConfigWithMultiController, ipAddress))
	s.NoError(s.InitMainController([]string{"--config=/tmp/k0s.yaml"}))

	token, err := s.GetJoinToken("controller", "")
	s.NoError(err)
	s.putFile("controller1", "/tmp/k0s.yaml", fmt.Sprintf(k0sConfigWithMultiController, ipAddress))
	s.NoError(s.JoinController(1, token, ""))

	s.putFile("controller2", "/tmp/k0s.yaml", fmt.Sprintf(k0sConfigWithMultiController, ipAddress))
	s.NoError(s.JoinController(2, token, ""))
	s.NoError(s.RunWorkers(""))

	kc, err := s.KubeClient("controller0", "")
	s.NoError(err)

	err = s.WaitForNodeReady("worker0", kc)
	s.NoError(err)

	pods, err := kc.CoreV1().Pods("kube-system").List(context.TODO(), v1.ListOptions{
		Limit: 100,
	})
	s.NoError(err)

	podCount := len(pods.Items)

	s.T().Logf("found %d pods in kube-system", podCount)
	s.Greater(podCount, 0, "expecting to see few pods in kube-system namespace")

	s.T().Log("waiting to see calico pods ready")
	s.NoError(common.WaitForCalicoReady(kc), "calico did not start")
}

func TestMultiControllerSuite(t *testing.T) {
	s := MultiControllerSuite{
		common.FootlooseSuite{
			ControllerCount: 3,
			WorkerCount:     1,
		},
	}
	suite.Run(t, &s)
}

func (s *MultiControllerSuite) putFile(node string, path string, content string) {
	ssh, err := s.SSH(node)
	s.Require().NoError(err)
	defer ssh.Disconnect()
	_, err = ssh.ExecWithOutput(fmt.Sprintf("echo '%s' >%s", content, path))

	s.Require().NoError(err)
}

const k0sConfigWithMultiController = `
spec:
  api:
    externalAddress: %s
`
