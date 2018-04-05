package rke

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/pki"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/stretchr/testify/assert"
)

var (
	dummyCertificate   *x509.Certificate
	dummyPrivateKey    *rsa.PrivateKey
	dummyPrivateKeyPEM string
)

const dummyCertPEM = `-----BEGIN CERTIFICATE-----
MIIDujCCAqKgAwIBAgIIE31FZVaPXTUwDQYJKoZIhvcNAQEFBQAwSTELMAkGA1UE
BhMCVVMxEzARBgNVBAoTCkdvb2dsZSBJbmMxJTAjBgNVBAMTHEdvb2dsZSBJbnRl
cm5ldCBBdXRob3JpdHkgRzIwHhcNMTQwMTI5MTMyNzQzWhcNMTQwNTI5MDAwMDAw
WjBpMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwN
TW91bnRhaW4gVmlldzETMBEGA1UECgwKR29vZ2xlIEluYzEYMBYGA1UEAwwPbWFp
bC5nb29nbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfRrObuSW5T7q
5CnSEqefEmtH4CCv6+5EckuriNr1CjfVvqzwfAhopXkLrq45EQm8vkmf7W96XJhC
7ZM0dYi1/qOCAU8wggFLMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAa
BgNVHREEEzARgg9tYWlsLmdvb2dsZS5jb20wCwYDVR0PBAQDAgeAMGgGCCsGAQUF
BwEBBFwwWjArBggrBgEFBQcwAoYfaHR0cDovL3BraS5nb29nbGUuY29tL0dJQUcy
LmNydDArBggrBgEFBQcwAYYfaHR0cDovL2NsaWVudHMxLmdvb2dsZS5jb20vb2Nz
cDAdBgNVHQ4EFgQUiJxtimAuTfwb+aUtBn5UYKreKvMwDAYDVR0TAQH/BAIwADAf
BgNVHSMEGDAWgBRK3QYWG7z2aLV29YG2u2IaulqBLzAXBgNVHSAEEDAOMAwGCisG
AQQB1nkCBQEwMAYDVR0fBCkwJzAloCOgIYYfaHR0cDovL3BraS5nb29nbGUuY29t
L0dJQUcyLmNybDANBgkqhkiG9w0BAQUFAAOCAQEAH6RYHxHdcGpMpFE3oxDoFnP+
gtuBCHan2yE2GRbJ2Cw8Lw0MmuKqHlf9RSeYfd3BXeKkj1qO6TVKwCh+0HdZk283
TZZyzmEOyclm3UGFYe82P/iDFt+CeQ3NpmBg+GoaVCuWAARJN/KfglbLyyYygcQq
0SgeDh8dRKUiaW3HQSoYvTvdTuqzwK4CXsr3b5/dAOY8uMuG/IAR3FgwTbZ1dtoW
RvOTa8hYiU6A475WuZKyEHcwnGYe57u2I2KbMgcKjPniocj4QzgYsVAVKW3IwaOh
yE+vPxsiUkvQHdO2fojCkY8jg70jxM+gu59tPDNbw3Uh/2Ij310FgTHsnGQMyA==
-----END CERTIFICATE-----
`

func init() {
	block, _ := pem.Decode([]byte(dummyCertPEM))
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}
	dummyCertificate = cert

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic("failed to generate rsa private key: " + err.Error())
	}
	dummyPrivateKey = key
	dummyPrivateKeyPEM = privateKeyToPEM(key)
}

type dummyResourceData struct {
	values map[string]interface{}
}

func (d *dummyResourceData) GetOk(key string) (interface{}, bool) {
	v, ok := d.values[key]
	return v, ok
}

type dummyStateBuilder struct {
	values map[string]interface{}
}

func (d *dummyStateBuilder) Set(k string, v interface{}) error {
	d.values[k] = v
	return nil
}
func (d *dummyStateBuilder) SetId(id string) { // nolint
	d.values["Id"] = id
}

func TestParseResourceRKEConfigNode(t *testing.T) {

	testcases := []struct {
		caseName     string
		resourceData map[string]interface{}
		expectNodes  []v3.RKEConfigNode
	}{
		{
			caseName: "minimum fields",
			resourceData: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{
						"address": "192.2.0.1",
						"role":    []interface{}{"etcd"},
					},
				},
			},
			expectNodes: []v3.RKEConfigNode{
				{
					Address: "192.2.0.1",
					Role:    []string{"etcd"},
				},
			},
		},
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{
						"node_name":         "node_name",
						"address":           "192.2.0.1",
						"port":              22,
						"internal_address":  "192.2.0.2",
						"role":              []interface{}{"controlplane", "worker", "etcd"},
						"hostname_override": "hostname_override",
						"user":              "rancher",
						"docker_socket":     "/var/run/docker.sock",
						"ssh_agent_auth":    true,
						"ssh_key":           "ssh_key",
						"ssh_key_path":      "ssh_key_path",
						"labels": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
					},
				},
			},
			expectNodes: []v3.RKEConfigNode{
				{
					NodeName:         "node_name",
					Address:          "192.2.0.1",
					Port:             "22",
					InternalAddress:  "192.2.0.2",
					Role:             []string{"controlplane", "worker", "etcd"},
					HostnameOverride: "hostname_override",
					User:             "rancher",
					DockerSocket:     "/var/run/docker.sock",
					SSHAgentAuth:     true,
					SSHKey:           "ssh_key",
					SSHKeyPath:       "ssh_key_path",
					Labels: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			nodes, err := parseResourceRKEConfigNode(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectNodes, nodes)
		})
	}
}

func TestParseResourceETCDService(t *testing.T) {
	testcases := []struct {
		caseName      string
		resourceData  map[string]interface{}
		expectService *v3.ETCDService
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"services_etcd": []interface{}{
					map[string]interface{}{
						"image": "image",
						"extra_args": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
						"extra_binds":   []interface{}{"/etc1", "/etc2"},
						"external_urls": []interface{}{"https://etcd1.example.com", "https://etcd2.example.com"},
						"ca_cert":       "ca_cert",
						"cert":          "cert",
						"key":           "key",
						"path":          "path",
					},
				},
			},
			expectService: &v3.ETCDService{
				BaseService: v3.BaseService{
					Image: "image",
					ExtraArgs: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
					ExtraBinds: []string{"/etc1", "/etc2"},
				},
				ExternalURLs: []string{"https://etcd1.example.com", "https://etcd2.example.com"},
				CACert:       "ca_cert",
				Cert:         "cert",
				Key:          "key",
				Path:         "path",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			service, err := parseResourceETCDService(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectService, service)
		})
	}
}

func TestParseResourceKubeAPIService(t *testing.T) {
	testcases := []struct {
		caseName      string
		resourceData  map[string]interface{}
		expectService *v3.KubeAPIService
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"services_kube_api": []interface{}{
					map[string]interface{}{
						"image": "image",
						"extra_args": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
						"extra_binds":              []interface{}{"/etc1", "/etc2"},
						"service_cluster_ip_range": "10.240.0.0/16",
						"pod_security_policy":      true,
					},
				},
			},
			expectService: &v3.KubeAPIService{
				BaseService: v3.BaseService{
					Image: "image",
					ExtraArgs: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
					ExtraBinds: []string{"/etc1", "/etc2"},
				},
				ServiceClusterIPRange: "10.240.0.0/16",
				PodSecurityPolicy:     true,
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			service, err := parseResourceKubeAPIService(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectService, service)
		})
	}
}

func TestParseResourceKubeControllerService(t *testing.T) {
	testcases := []struct {
		caseName      string
		resourceData  map[string]interface{}
		expectService *v3.KubeControllerService
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"services_kube_controller": []interface{}{
					map[string]interface{}{
						"image": "image",
						"extra_args": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
						"extra_binds":              []interface{}{"/etc1", "/etc2"},
						"cluster_cidr":             "10.240.0.0/16",
						"service_cluster_ip_range": "10.240.0.0/16",
					},
				},
			},
			expectService: &v3.KubeControllerService{
				BaseService: v3.BaseService{
					Image: "image",
					ExtraArgs: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
					ExtraBinds: []string{"/etc1", "/etc2"},
				},
				ClusterCIDR:           "10.240.0.0/16",
				ServiceClusterIPRange: "10.240.0.0/16",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			service, err := parseResourceKubeControllerService(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectService, service)
		})
	}
}

func TestParseResourceSchedulerService(t *testing.T) {
	testcases := []struct {
		caseName      string
		resourceData  map[string]interface{}
		expectService *v3.SchedulerService
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"services_scheduler": []interface{}{
					map[string]interface{}{
						"image": "image",
						"extra_args": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
						"extra_binds": []interface{}{"/etc1", "/etc2"},
					},
				},
			},
			expectService: &v3.SchedulerService{
				BaseService: v3.BaseService{
					Image: "image",
					ExtraArgs: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
					ExtraBinds: []string{"/etc1", "/etc2"},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			service, err := parseResourceSchedulerService(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectService, service)
		})
	}
}

func TestParseResourceKubeletService(t *testing.T) {
	testcases := []struct {
		caseName      string
		resourceData  map[string]interface{}
		expectService *v3.KubeletService
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"services_kubelet": []interface{}{
					map[string]interface{}{
						"image": "image",
						"extra_args": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
						"extra_binds":           []interface{}{"/etc1", "/etc2"},
						"cluster_domain":        "example.com",
						"infra_container_image": "alpine:latest",
						"cluster_dns_server":    "192.2.0.1",
						"fail_swap_on":          true,
					},
				},
			},
			expectService: &v3.KubeletService{
				BaseService: v3.BaseService{
					Image: "image",
					ExtraArgs: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
					ExtraBinds: []string{"/etc1", "/etc2"},
				},
				ClusterDomain:       "example.com",
				InfraContainerImage: "alpine:latest",
				ClusterDNSServer:    "192.2.0.1",
				FailSwapOn:          true,
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			service, err := parseResourceKubeletService(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectService, service)
		})
	}
}

func TestParseResourceKubeproxyService(t *testing.T) {
	testcases := []struct {
		caseName      string
		resourceData  map[string]interface{}
		expectService *v3.KubeproxyService
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"services_kubeproxy": []interface{}{
					map[string]interface{}{
						"image": "image",
						"extra_args": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
						"extra_binds": []interface{}{"/etc1", "/etc2"},
					},
				},
			},
			expectService: &v3.KubeproxyService{
				BaseService: v3.BaseService{
					Image: "image",
					ExtraArgs: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
					ExtraBinds: []string{"/etc1", "/etc2"},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			service, err := parseResourceKubeproxyService(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectService, service)
		})
	}
}

func TestParseResourceNetwork(t *testing.T) {
	testcases := []struct {
		caseName      string
		resourceData  map[string]interface{}
		expectNetwork *v3.NetworkConfig
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"network": []interface{}{
					map[string]interface{}{
						"plugin": "calico",
						"options": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
					},
				},
			},
			expectNetwork: &v3.NetworkConfig{
				Plugin: "calico",
				Options: map[string]string{
					"foo": "foo",
					"bar": "bar",
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			network, err := parseResourceNetwork(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectNetwork, network)
		})
	}
}

func TestParseResourceAuthentication(t *testing.T) {
	testcases := []struct {
		caseName     string
		resourceData map[string]interface{}
		expectConfig *v3.AuthnConfig
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"authentication": []interface{}{
					map[string]interface{}{
						"strategy": "x509",
						"options": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
						"sans": []interface{}{
							"192.2.0.1",
							"test.example.com",
						},
					},
				},
			},
			expectConfig: &v3.AuthnConfig{
				Strategy: "x509",
				Options: map[string]string{
					"foo": "foo",
					"bar": "bar",
				},
				SANs: []string{
					"192.2.0.1",
					"test.example.com",
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			config, err := parseResourceAuthentication(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectConfig, config)
		})
	}
}

func TestParseResourceAddons(t *testing.T) {
	d := &dummyResourceData{values: map[string]interface{}{"addons": "addons: yaml"}}
	addon, err := parseResourceAddons(d)
	assert.NoError(t, err)
	assert.EqualValues(t, "addons: yaml", addon)
}

func TestParseResourceAddonsInclude(t *testing.T) {
	expect := []string{
		"https://example.com/addon1.yaml",
		"https://example.com/addon2.yaml",
	}
	d := &dummyResourceData{
		values: map[string]interface{}{
			"addons_include": []interface{}{
				"https://example.com/addon1.yaml",
				"https://example.com/addon2.yaml",
			},
		},
	}
	includes, err := parseResourceAddonsInclude(d)
	assert.NoError(t, err)
	assert.EqualValues(t, expect, includes)
}

func TestParseResourceSystemImages(t *testing.T) {
	testcases := []struct {
		caseName     string
		resourceData map[string]interface{}
		expectConfig *v3.RKESystemImages
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"system_images": []interface{}{
					map[string]interface{}{
						"etcd":                        "etcd",
						"alpine":                      "alpine",
						"nginx_proxy":                 "nginx_proxy",
						"cert_downloader":             "cert_downloader",
						"kubernetes_services_sidecar": "kubernetes_services_sidecar",
						"kube_dns":                    "kube_dns",
						"dnsmasq":                     "dnsmasq",
						"kube_dns_sidecar":            "kube_dns_sidecar",
						"kube_dns_autoscaler":         "kube_dns_autoscaler",
						"kubernetes":                  "kubernetes",
						"flannel":                     "flannel",
						"flannel_cni":                 "flannel_cni",
						"calico_node":                 "calico_node",
						"calico_cni":                  "calico_cni",
						"calico_controllers":          "calico_controllers",
						"calico_ctl":                  "calico_ctl",
						"canal_node":                  "canal_node",
						"canal_cni":                   "canal_cni",
						"canal_flannel":               "canal_flannel",
						"weave_node":                  "weave_node",
						"weave_cni":                   "weave_cni",
						"pod_infra_container":         "pod_infra_container",
						"ingress":                     "ingress",
						"ingress_backend":             "ingress_backend",
						"dashboard":                   "dashboard",
						"heapster":                    "heapster",
						"grafana":                     "grafana",
						"influxdb":                    "influxdb",
						"tiller":                      "tiller",
					},
				},
			},
			expectConfig: &v3.RKESystemImages{
				Etcd:                      "etcd",
				Alpine:                    "alpine",
				NginxProxy:                "nginx_proxy",
				CertDownloader:            "cert_downloader",
				KubernetesServicesSidecar: "kubernetes_services_sidecar",
				KubeDNS:                   "kube_dns",
				DNSmasq:                   "dnsmasq",
				KubeDNSSidecar:            "kube_dns_sidecar",
				KubeDNSAutoscaler:         "kube_dns_autoscaler",
				Kubernetes:                "kubernetes",
				Flannel:                   "flannel",
				FlannelCNI:                "flannel_cni",
				CalicoNode:                "calico_node",
				CalicoCNI:                 "calico_cni",
				CalicoControllers:         "calico_controllers",
				CalicoCtl:                 "calico_ctl",
				CanalNode:                 "canal_node",
				CanalCNI:                  "canal_cni",
				CanalFlannel:              "canal_flannel",
				WeaveNode:                 "weave_node",
				WeaveCNI:                  "weave_cni",
				PodInfraContainer:         "pod_infra_container",
				Ingress:                   "ingress",
				IngressBackend:            "ingress_backend",
				Dashboard:                 "dashboard",
				Heapster:                  "heapster",
				Grafana:                   "grafana",
				Influxdb:                  "influxdb",
				Tiller:                    "tiller",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			config, err := parseResourceSystemImages(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectConfig, config)
		})
	}
}

func TestParseResourceSSHKeyPath(t *testing.T) {
	d := &dummyResourceData{values: map[string]interface{}{"ssh_key_path": "ssh_key_path"}}
	keyPath, err := parseResourceSSHKeyPath(d)
	assert.NoError(t, err)
	assert.EqualValues(t, "ssh_key_path", keyPath)
}

func TestParseResourceSSHAgentAuth(t *testing.T) {
	d := &dummyResourceData{values: map[string]interface{}{"ssh_agent_auth": true}}
	auth, err := parseResourceSSHAgentAuth(d)
	assert.NoError(t, err)
	assert.EqualValues(t, true, auth)
}

func TestParseResourceAuthorization(t *testing.T) {
	testcases := []struct {
		caseName     string
		resourceData map[string]interface{}
		expectConfig *v3.AuthzConfig
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"authorization": []interface{}{
					map[string]interface{}{
						"mode": "rbac",
						"options": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
					},
				},
			},
			expectConfig: &v3.AuthzConfig{
				Mode: "rbac",
				Options: map[string]string{
					"foo": "foo",
					"bar": "bar",
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			config, err := parseResourceAuthorization(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectConfig, config)
		})
	}
}

func TestParseResourceIgnoreDockerVersion(t *testing.T) {
	d := &dummyResourceData{values: map[string]interface{}{"ignore_docker_version": true}}
	ignore, err := parseResourceIgnoreDockerVersion(d)
	assert.NoError(t, err)
	assert.EqualValues(t, true, ignore)
}

func TestParseResourceKubernetesVersion(t *testing.T) {
	d := &dummyResourceData{
		values: map[string]interface{}{
			"kubernetes_version": "1.8.9",
		},
	}
	version, err := parseResourceVersion(d)
	assert.NoError(t, err)
	assert.EqualValues(t, "1.8.9", version)
}

func TestParseResourcePrivateRegistries(t *testing.T) {
	testcases := []struct {
		caseName     string
		resourceData map[string]interface{}
		expectConfig []v3.PrivateRegistry
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"private_registries": []interface{}{
					map[string]interface{}{
						"url":      "https://example.com",
						"user":     "rancher",
						"password": "p@ssword",
					},
				},
			},
			expectConfig: []v3.PrivateRegistry{
				{
					URL:      "https://example.com",
					User:     "rancher",
					Password: "p@ssword",
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			config, err := parseResourcePrivateRegistries(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectConfig, config)
		})
	}
}

func TestParseResourceIngress(t *testing.T) {
	testcases := []struct {
		caseName     string
		resourceData map[string]interface{}
		expectConfig *v3.IngressConfig
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"ingress": []interface{}{
					map[string]interface{}{
						"provider": "nginx",
						"options": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
						"node_selector": map[string]interface{}{
							"role": "worker",
						},
					},
				},
			},
			expectConfig: &v3.IngressConfig{
				Provider: "nginx",
				Options: map[string]string{
					"foo": "foo",
					"bar": "bar",
				},
				NodeSelector: map[string]string{
					"role": "worker",
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			config, err := parseResourceIngress(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectConfig, config)
		})
	}
}

func TestParseResourceClusterName(t *testing.T) {
	d := &dummyResourceData{
		values: map[string]interface{}{
			"cluster_name": "rke",
		},
	}
	name, err := parseResourceClusterName(d)
	assert.NoError(t, err)
	assert.EqualValues(t, "rke", name)
}

func TestParseResourceCloudProvider(t *testing.T) {
	testcases := []struct {
		caseName     string
		resourceData map[string]interface{}
		expectConfig *v3.CloudProvider
	}{
		{
			caseName: "all fields",
			resourceData: map[string]interface{}{
				"cloud_provider": []interface{}{
					map[string]interface{}{
						"name": "sakuracloud",
						"cloud_config": map[string]interface{}{
							"token":  "your-token",
							"secret": "your-secret",
							"zone":   "your-zone",
						},
					},
				},
			},
			expectConfig: &v3.CloudProvider{
				Name: "sakuracloud",
				CloudConfig: map[string]string{
					"token":  "your-token",
					"secret": "your-secret",
					"zone":   "your-zone",
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyResourceData{values: testcase.resourceData}
			config, err := parseResourceCloudProvider(d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.expectConfig, config)
		})
	}
}

func TestClusterToState(t *testing.T) {

	testcases := []struct {
		caseName string
		cluster  *cluster.Cluster
		state    map[string]interface{}
	}{
		{
			caseName: "all fields",
			cluster: &cluster.Cluster{
				RancherKubernetesEngineConfig: v3.RancherKubernetesEngineConfig{
					Nodes: []v3.RKEConfigNode{
						{
							NodeName:         "node_name",
							Address:          "192.2.0.1",
							Port:             "22",
							InternalAddress:  "192.2.0.2",
							Role:             []string{"role1", "role2"},
							HostnameOverride: "hostname_override",
							User:             "rancher",
							DockerSocket:     "/var/run/docker.sock",
							SSHAgentAuth:     true,
							SSHKey:           "ssh_key",
							SSHKeyPath:       "ssh_key_path",
							Labels: map[string]string{
								"foo": "foo",
								"bar": "bar",
							},
						},
					},
					Services: v3.RKEConfigServices{
						Etcd: v3.ETCDService{
							BaseService: v3.BaseService{
								Image: "etcd:latest",
								ExtraArgs: map[string]string{
									"foo": "bar",
									"bar": "foo",
								},
								ExtraBinds: []string{"/bind1", "/bind2"},
							},
							ExternalURLs: []string{
								"https://ext1.example.com",
								"https://ext2.example.com",
							},
							CACert: "ca_cert",
							Cert:   "cert",
							Key:    "key",
							Path:   "path",
						},
						KubeAPI: v3.KubeAPIService{
							BaseService: v3.BaseService{
								Image: "kube_api:latest",
								ExtraArgs: map[string]string{
									"foo": "bar",
									"bar": "foo",
								},
								ExtraBinds: []string{"/bind1", "/bind2"},
							},
							ServiceClusterIPRange: "10.240.0.0/16",
							PodSecurityPolicy:     true,
						},
						KubeController: v3.KubeControllerService{
							BaseService: v3.BaseService{
								Image: "kube_controller:latest",
								ExtraArgs: map[string]string{
									"foo": "bar",
									"bar": "foo",
								},
								ExtraBinds: []string{"/bind1", "/bind2"},
							},
							ClusterCIDR:           "10.200.0.0/8",
							ServiceClusterIPRange: "10.240.0.0/16",
						},
						Scheduler: v3.SchedulerService{
							BaseService: v3.BaseService{
								Image: "scheduler:latest",
								ExtraArgs: map[string]string{
									"foo": "bar",
									"bar": "foo",
								},
								ExtraBinds: []string{"/bind1", "/bind2"},
							},
						},
						Kubelet: v3.KubeletService{
							BaseService: v3.BaseService{
								Image: "kubelet:latest",
								ExtraArgs: map[string]string{
									"foo": "bar",
									"bar": "foo",
								},
								ExtraBinds: []string{"/bind1", "/bind2"},
							},
							ClusterDomain:       "example.com",
							InfraContainerImage: "alpine:latest",
							ClusterDNSServer:    "192.2.0.1",
							FailSwapOn:          true,
						},
						Kubeproxy: v3.KubeproxyService{
							BaseService: v3.BaseService{
								Image: "kubeproxy:latest",
								ExtraArgs: map[string]string{
									"foo": "bar",
									"bar": "foo",
								},
								ExtraBinds: []string{"/bind1", "/bind2"},
							},
						},
					},
					Network: v3.NetworkConfig{
						Plugin: "calico",
						Options: map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
					},
					Authentication: v3.AuthnConfig{
						Strategy: "x509",
						Options: map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						SANs: []string{"sans1", "sans2"},
					},
					Addons: "addons: yaml",
					AddonsInclude: []string{
						"https://example.com/addon1.yaml",
						"https://example.com/addon2.yaml",
					},
					SystemImages: v3.RKESystemImages{
						Etcd:                      "etcd",
						Alpine:                    "alpine",
						NginxProxy:                "nginx_proxy",
						CertDownloader:            "cert_downloader",
						KubernetesServicesSidecar: "kubernetes_services_sidecar",
						KubeDNS:                   "kube_dns",
						DNSmasq:                   "dnsmasq",
						KubeDNSSidecar:            "kube_dns_sidecar",
						KubeDNSAutoscaler:         "kube_dns_autoscaler",
						Kubernetes:                "kubernetes",
						Flannel:                   "flannel",
						FlannelCNI:                "flannel_cni",
						CalicoNode:                "calico_node",
						CalicoCNI:                 "calico_cni",
						CalicoControllers:         "calico_controllers",
						CalicoCtl:                 "calico_ctl",
						CanalNode:                 "canal_node",
						CanalCNI:                  "canal_cni",
						CanalFlannel:              "canal_flannel",
						WeaveNode:                 "weave_node",
						WeaveCNI:                  "weave_cni",
						PodInfraContainer:         "pod_infra_container",
						Ingress:                   "ingress",
						IngressBackend:            "ingress_backend",
						Dashboard:                 "dashboard",
						Heapster:                  "heapster",
						Grafana:                   "grafana",
						Influxdb:                  "influxdb",
						Tiller:                    "tiller",
					},
					SSHKeyPath:   "ssh_key_path",
					SSHAgentAuth: true,
					Authorization: v3.AuthzConfig{
						Mode: "rbac",
						Options: map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
					},
					IgnoreDockerVersion: true,
					Version:             "1.8.9",
					PrivateRegistries: []v3.PrivateRegistry{
						{
							URL:      "https://registry1.example.com",
							User:     "user1",
							Password: "password1",
						},
						{
							URL:      "https://registry2.example.com",
							User:     "user2",
							Password: "password2",
						},
					},
					Ingress: v3.IngressConfig{
						Provider: "nginx",
						Options: map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						NodeSelector: map[string]string{
							"role": "worker",
						},
					},
					ClusterName: "example",
					CloudProvider: v3.CloudProvider{
						Name: "sakuracloud",
						CloudConfig: map[string]string{
							"token":  "your-token",
							"secret": "your-secret",
							"zone":   "your-zone",
						},
					},
				},
				EtcdHosts: []*hosts.Host{
					{
						RKEConfigNode: v3.RKEConfigNode{
							NodeName: "etcd1",
							Address:  "192.2.0.1",
						},
					},
					{
						RKEConfigNode: v3.RKEConfigNode{
							NodeName: "etcd2",
							Address:  "192.2.0.2",
						},
					},
				},
				WorkerHosts: []*hosts.Host{
					{
						RKEConfigNode: v3.RKEConfigNode{
							NodeName: "host",
							Address:  "192.2.0.1",
						},
					},
				},
				ControlPlaneHosts: []*hosts.Host{
					{
						RKEConfigNode: v3.RKEConfigNode{
							NodeName: "host",
							Address:  "192.2.0.1",
						},
					},
				},
				InactiveHosts: []*hosts.Host{
					{
						RKEConfigNode: v3.RKEConfigNode{
							NodeName: "host",
							Address:  "192.2.0.1",
						},
					},
				},
				Certificates: map[string]pki.CertificatePKI{
					"example": {
						Certificate:   dummyCertificate,
						Key:           dummyPrivateKey,
						Config:        "config",
						Name:          "name",
						CommonName:    "common_name",
						OUName:        "ou_name",
						EnvName:       "env_name",
						Path:          "path",
						KeyEnvName:    "key_env_name",
						KeyPath:       "key_path",
						ConfigEnvName: "config_env_name",
						ConfigPath:    "config_path",
					},
				},
				ClusterDomain:    "example.com",
				ClusterCIDR:      "10.200.0.0/8",
				ClusterDNSServer: "192.2.0.1",
			},
			state: map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{
						"node_name":         "node_name",
						"address":           "192.2.0.1",
						"port":              22,
						"internal_address":  "192.2.0.2",
						"role":              []string{"role1", "role2"},
						"hostname_override": "hostname_override",
						"user":              "rancher",
						"docker_socket":     "/var/run/docker.sock",
						"ssh_agent_auth":    true,
						"ssh_key":           "ssh_key",
						"ssh_key_path":      "ssh_key_path",
						"labels": map[string]string{
							"foo": "foo",
							"bar": "bar",
						},
					},
				},
				"services_etcd": []interface{}{
					map[string]interface{}{
						"image": "etcd:latest",
						"extra_args": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						"extra_binds": []string{"/bind1", "/bind2"},
						"external_urls": []string{
							"https://ext1.example.com",
							"https://ext2.example.com",
						},
						"ca_cert": "ca_cert",
						"cert":    "cert",
						"key":     "key",
						"path":    "path",
					},
				},
				"services_kube_api": []interface{}{
					map[string]interface{}{
						"image": "kube_api:latest",
						"extra_args": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						"extra_binds":              []string{"/bind1", "/bind2"},
						"service_cluster_ip_range": "10.240.0.0/16",
						"pod_security_policy":      true,
					},
				},
				"services_kube_controller": []interface{}{
					map[string]interface{}{
						"image": "kube_controller:latest",
						"extra_args": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						"extra_binds":              []string{"/bind1", "/bind2"},
						"cluster_cidr":             "10.200.0.0/8",
						"service_cluster_ip_range": "10.240.0.0/16",
					},
				},
				"services_scheduler": []interface{}{
					map[string]interface{}{
						"image": "scheduler:latest",
						"extra_args": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						"extra_binds": []string{"/bind1", "/bind2"},
					},
				},
				"services_kubelet": []interface{}{
					map[string]interface{}{
						"image": "kubelet:latest",
						"extra_args": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						"extra_binds":           []string{"/bind1", "/bind2"},
						"cluster_domain":        "example.com",
						"infra_container_image": "alpine:latest",
						"cluster_dns_server":    "192.2.0.1",
						"fail_swap_on":          true,
					},
				},
				"services_kubeproxy": []interface{}{
					map[string]interface{}{
						"image": "kubeproxy:latest",
						"extra_args": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						"extra_binds": []string{"/bind1", "/bind2"},
					},
				},
				"network": []interface{}{
					map[string]interface{}{
						"plugin": "calico",
						"options": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
					},
				},
				"authentication": []interface{}{
					map[string]interface{}{
						"strategy": "x509",
						"options": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						"sans": []string{"sans1", "sans2"},
					},
				},
				"addons": "addons: yaml",
				"addons_include": []string{
					"https://example.com/addon1.yaml",
					"https://example.com/addon2.yaml",
				},
				"system_images": []interface{}{
					map[string]interface{}{
						"etcd":                        "etcd",
						"alpine":                      "alpine",
						"nginx_proxy":                 "nginx_proxy",
						"cert_downloader":             "cert_downloader",
						"kubernetes_services_sidecar": "kubernetes_services_sidecar",
						"kube_dns":                    "kube_dns",
						"dnsmasq":                     "dnsmasq",
						"kube_dns_sidecar":            "kube_dns_sidecar",
						"kube_dns_autoscaler":         "kube_dns_autoscaler",
						"kubernetes":                  "kubernetes",
						"flannel":                     "flannel",
						"flannel_cni":                 "flannel_cni",
						"calico_node":                 "calico_node",
						"calico_cni":                  "calico_cni",
						"calico_controllers":          "calico_controllers",
						"calico_ctl":                  "calico_ctl",
						"canal_node":                  "canal_node",
						"canal_cni":                   "canal_cni",
						"canal_flannel":               "canal_flannel",
						"weave_node":                  "weave_node",
						"weave_cni":                   "weave_cni",
						"pod_infra_container":         "pod_infra_container",
						"ingress":                     "ingress",
						"ingress_backend":             "ingress_backend",
						"dashboard":                   "dashboard",
						"heapster":                    "heapster",
						"grafana":                     "grafana",
						"influxdb":                    "influxdb",
						"tiller":                      "tiller",
					},
				},
				"ssh_key_path":   "ssh_key_path",
				"ssh_agent_auth": true,
				"authorization": []interface{}{
					map[string]interface{}{
						"mode": "rbac",
						"options": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
					},
				},
				"ignore_docker_version": true,
				"kubernetes_version":    "1.8.9",
				"private_registries": []interface{}{
					map[string]interface{}{
						"url":      "https://registry1.example.com",
						"user":     "user1",
						"password": "password1",
					},
					map[string]interface{}{
						"url":      "https://registry2.example.com",
						"user":     "user2",
						"password": "password2",
					},
				},
				"ingress": []interface{}{
					map[string]interface{}{
						"provider": "nginx",
						"options": map[string]string{
							"foo": "bar",
							"bar": "foo",
						},
						"node_selector": map[string]string{
							"role": "worker",
						},
					},
				},
				"cluster_name": "example",
				"cloud_provider": []interface{}{
					map[string]interface{}{
						"name": "sakuracloud",
						"cloud_config": map[string]string{
							"token":  "your-token",
							"secret": "your-secret",
							"zone":   "your-zone",
						},
					},
				},
				"certificates": []interface{}{
					map[string]interface{}{
						"id":              "example",
						"certificate":     dummyCertPEM,
						"key":             dummyPrivateKeyPEM,
						"config":          "config",
						"name":            "name",
						"common_name":     "common_name",
						"ou_name":         "ou_name",
						"env_name":        "env_name",
						"path":            "path",
						"key_env_name":    "key_env_name",
						"key_path":        "key_path",
						"config_env_name": "config_env_name",
						"config_path":     "config_path",
					},
				},
				"cluster_domain":     "example.com",
				"cluster_cidr":       "10.200.0.0/8",
				"cluster_dns_server": "192.2.0.1",
				"etcd_hosts": []map[string]interface{}{
					{
						"node_name": "etcd1",
						"address":   "192.2.0.1",
					},
					{
						"node_name": "etcd2",
						"address":   "192.2.0.2",
					},
				},
				"worker_hosts": []map[string]interface{}{
					{
						"node_name": "host",
						"address":   "192.2.0.1",
					},
				},
				"control_plane_hosts": []map[string]interface{}{
					{
						"node_name": "host",
						"address":   "192.2.0.1",
					},
				},
				"inactive_hosts": []map[string]interface{}{
					{
						"node_name": "host",
						"address":   "192.2.0.1",
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.caseName, func(t *testing.T) {
			d := &dummyStateBuilder{values: map[string]interface{}{}}
			err := clusterToState(testcase.cluster, d)
			assert.NoError(t, err)
			assert.EqualValues(t, testcase.state, d.values)
		})
	}

}