package hclutils_test

import (
	"testing"

	"github.com/hashicorp/hcl2/hcldec"
	"github.com/hashicorp/nomad/drivers/docker"
	"github.com/hashicorp/nomad/helper/pluginutils/hclspecutils"
	"github.com/hashicorp/nomad/helper/pluginutils/hclutils"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestParseHclInterface_Hcl(t *testing.T) {
	dockerDriver := new(docker.Driver)
	dockerSpec, err := dockerDriver.TaskConfigSchema()
	require.NoError(t, err)
	dockerDecSpec, diags := hclspecutils.Convert(dockerSpec)
	require.False(t, diags.HasErrors())

	vars := map[string]cty.Value{
		"NOMAD_ALLOC_INDEX": cty.NumberIntVal(2),
		"NOMAD_META_hello":  cty.StringVal("world"),
	}

	cases := []struct {
		name         string
		config       interface{}
		spec         hcldec.Spec
		vars         map[string]cty.Value
		expected     interface{}
		expectedType interface{}
	}{
		{
			name: "single string attr",
			config: hclutils.HclConfigToInterface(t, `
			config {
				image = "redis:3.2"
			}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:   "redis:3.2",
				Devices: []docker.DockerDevice{},
				Mounts:  []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "single string attr json",
			config: hclutils.JsonConfigToInterface(t, `
						{
							"Config": {
								"image": "redis:3.2"
			                }
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:   "redis:3.2",
				Devices: []docker.DockerDevice{},
				Mounts:  []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "number attr",
			config: hclutils.HclConfigToInterface(t, `
						config {
							image = "redis:3.2"
							pids_limit  = 2
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:     "redis:3.2",
				PidsLimit: 2,
				Devices:   []docker.DockerDevice{},
				Mounts:    []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "number attr json",
			config: hclutils.JsonConfigToInterface(t, `
						{
							"Config": {
								"image": "redis:3.2",
								"pids_limit": "2"
			                }
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:     "redis:3.2",
				PidsLimit: 2,
				Devices:   []docker.DockerDevice{},
				Mounts:    []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "number attr interpolated",
			config: hclutils.HclConfigToInterface(t, `
						config {
							image = "redis:3.2"
							pids_limit  = "${2 + 2}"
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:     "redis:3.2",
				PidsLimit: 4,
				Devices:   []docker.DockerDevice{},
				Mounts:    []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "number attr interploated json",
			config: hclutils.JsonConfigToInterface(t, `
						{
							"Config": {
								"image": "redis:3.2",
								"pids_limit": "${2 + 2}"
			                }
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:     "redis:3.2",
				PidsLimit: 4,
				Devices:   []docker.DockerDevice{},
				Mounts:    []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "multi attr",
			config: hclutils.HclConfigToInterface(t, `
						config {
							image = "redis:3.2"
							args = ["foo", "bar"]
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:   "redis:3.2",
				Args:    []string{"foo", "bar"},
				Devices: []docker.DockerDevice{},
				Mounts:  []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "multi attr json",
			config: hclutils.JsonConfigToInterface(t, `
						{
							"Config": {
								"image": "redis:3.2",
								"args": ["foo", "bar"]
			                }
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:   "redis:3.2",
				Args:    []string{"foo", "bar"},
				Devices: []docker.DockerDevice{},
				Mounts:  []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "multi attr variables",
			config: hclutils.HclConfigToInterface(t, `
						config {
							image = "redis:3.2"
							args = ["${NOMAD_META_hello}", "${NOMAD_ALLOC_INDEX}"]
							pids_limit = "${NOMAD_ALLOC_INDEX + 2}"
						}`),
			spec: dockerDecSpec,
			vars: vars,
			expected: &docker.TaskConfig{
				Image:     "redis:3.2",
				Args:      []string{"world", "2"},
				PidsLimit: 4,
				Devices:   []docker.DockerDevice{},
				Mounts:    []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "multi attr variables json",
			config: hclutils.JsonConfigToInterface(t, `
						{
							"Config": {
								"image": "redis:3.2",
								"args": ["foo", "bar"]
			                }
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:   "redis:3.2",
				Args:    []string{"foo", "bar"},
				Devices: []docker.DockerDevice{},
				Mounts:  []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "port_map",
			config: hclutils.HclConfigToInterface(t, `
			config {
				image = "redis:3.2"
				port_map {
					foo = 1234
					bar = 5678
				}
			}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image: "redis:3.2",
				PortMap: map[string]int{
					"foo": 1234,
					"bar": 5678,
				},
				Devices: []docker.DockerDevice{},
				Mounts:  []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "port_map json",
			config: hclutils.JsonConfigToInterface(t, `
							{
								"Config": {
									"image": "redis:3.2",
									"port_map": [{
										"foo": 1234,
										"bar": 5678
									}]
				                }
							}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image: "redis:3.2",
				PortMap: map[string]int{
					"foo": 1234,
					"bar": 5678,
				},
				Devices: []docker.DockerDevice{},
				Mounts:  []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "devices",
			config: hclutils.HclConfigToInterface(t, `
						config {
							image = "redis:3.2"
							devices = [
								{
									host_path = "/dev/sda1"
									container_path = "/dev/xvdc"
									cgroup_permissions = "r"
								},
								{
									host_path = "/dev/sda2"
									container_path = "/dev/xvdd"
								}
							]
						}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image: "redis:3.2",
				Devices: []docker.DockerDevice{
					{
						HostPath:          "/dev/sda1",
						ContainerPath:     "/dev/xvdc",
						CgroupPermissions: "r",
					},
					{
						HostPath:      "/dev/sda2",
						ContainerPath: "/dev/xvdd",
					},
				},
				Mounts: []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "docker_logging",
			config: hclutils.HclConfigToInterface(t, `
				config {
					image = "redis:3.2"
					network_mode = "host"
					dns_servers = ["169.254.1.1"]
					logging {
					    type = "syslog"
					    config {
						tag  = "driver-test"
					    }
					}
				}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image:       "redis:3.2",
				NetworkMode: "host",
				DNSServers:  []string{"169.254.1.1"},
				Logging: docker.DockerLogging{
					Type: "syslog",
					Config: map[string]string{
						"tag": "driver-test",
					},
				},
				Devices: []docker.DockerDevice{},
				Mounts:  []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
		{
			name: "docker_json",
			config: hclutils.JsonConfigToInterface(t, `
					{
						"Config": {
							"image": "redis:3.2",
							"devices": [
								{
									"host_path": "/dev/sda1",
									"container_path": "/dev/xvdc",
									"cgroup_permissions": "r"
								},
								{
									"host_path": "/dev/sda2",
									"container_path": "/dev/xvdd"
								}
							]
				}
					}`),
			spec: dockerDecSpec,
			expected: &docker.TaskConfig{
				Image: "redis:3.2",
				Devices: []docker.DockerDevice{
					{
						HostPath:          "/dev/sda1",
						ContainerPath:     "/dev/xvdc",
						CgroupPermissions: "r",
					},
					{
						HostPath:      "/dev/sda2",
						ContainerPath: "/dev/xvdd",
					},
				},
				Mounts: []docker.DockerMount{},
			},
			expectedType: &docker.TaskConfig{},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Logf("Val: % #v", pretty.Formatter(c.config))
			// Parse the interface
			ctyValue, diag := hclutils.ParseHclInterface(c.config, c.spec, c.vars)
			if diag.HasErrors() {
				for _, err := range diag.Errs() {
					t.Error(err)
				}
				t.FailNow()
			}

			// Test encoding
			taskConfig := &drivers.TaskConfig{}
			require.NoError(t, taskConfig.EncodeDriverConfig(ctyValue))

			// Test decoding
			require.NoError(t, taskConfig.DecodeDriverConfig(c.expectedType))

			require.EqualValues(t, c.expected, c.expectedType)

		})
	}
}

func TestParseNullFields(t *testing.T) {
	spec := hclspec.NewObject(map[string]*hclspec.Spec{
		"array_field":   hclspec.NewAttr("array_field", "list(string)", false),
		"string_field":  hclspec.NewAttr("string_field", "string", false),
		"boolean_field": hclspec.NewAttr("boolean_field", "bool", false),
		"number_field":  hclspec.NewAttr("number_field", "number", false),
		"block_field": hclspec.NewBlock("block_field", false, hclspec.NewObject((map[string]*hclspec.Spec{
			"f": hclspec.NewAttr("f", "string", true),
		}))),
		"block_list_field": hclspec.NewBlockList("block_list_field", hclspec.NewObject((map[string]*hclspec.Spec{
			"f": hclspec.NewAttr("f", "string", true),
		}))),
	})

	type Sub struct {
		F string `codec:"f"`
	}

	type TaskConfig struct {
		Array     []string `codec:"array_field"`
		String    string   `codec:"string_field"`
		Boolean   bool     `codec:"boolean_field"`
		Number    int64    `codec:"number_field"`
		Block     Sub      `codec:"block_field"`
		BlockList []Sub    `codec:"block_list_field"`
	}

	cases := []struct {
		name     string
		json     string
		expected TaskConfig
	}{
		{
			"omitted fields",
			`{"Config": {}}`,
			TaskConfig{BlockList: []Sub{}},
		},
		{
			"explicitly nil",
			`{"Config": {
                            "array_field": null,
                            "string_field": null,
			    "boolean_field": null,
                            "number_field": null,
                            "block_field": null,
                            "block_list_field": null}}`,
			TaskConfig{BlockList: []Sub{}},
		},
		{
			// for sanity checking that the fields are actually set
			"explicitly set to not null",
			`{"Config": {
                            "array_field": ["a"],
                            "string_field": "a",
                            "boolean_field": true,
                            "number_field": 5,
                            "block_field": [{"f": "a"}],
                            "block_list_field": [{"f": "a"}, {"f": "b"}]}}`,
			TaskConfig{
				Array:     []string{"a"},
				String:    "a",
				Boolean:   true,
				Number:    5,
				Block:     Sub{"a"},
				BlockList: []Sub{{"a"}, {"b"}},
			},
		},
	}

	parser := hclutils.NewConfigParser(spec)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var tc TaskConfig
			parser.ParseJson(t, c.json, &tc)

			require.EqualValues(t, c.expected, tc)
		})
	}
}
