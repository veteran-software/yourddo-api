package main

import (
	"strings"
	"testing"
)

func TestParseDatacenterXML(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		want    *ArrayOfDatacenterStruct
		wantErr bool
	}{
		{
			name: "valid datacenter XML",
			xml: `<ArrayOfDatacenterStruct>
				<DatacenterStruct>
					<KeyName>TestDC</KeyName>
					<Datacenter>
						<cachedAt>2024-01-01T00:00:00Z</cachedAt>
						<datacenter>
							<Datacenter>
								<Name>Test Datacenter</Name>
								<AuthServer>auth.test.com</AuthServer>
								<PatchServer>patch.test.com</PatchServer>
								<LauncherConfigurationServer>launcher.test.com</LauncherConfigurationServer>
								<Worlds>
									<World>
										<Name>TestWorld</Name>
										<LoginServerUrl>login.test.com</LoginServerUrl>
										<ChatServerUrl>chat.test.com</ChatServerUrl>
										<StatusServerUrl>status.test.com</StatusServerUrl>
										<Order>1</Order>
									</World>
								</Worlds>
							</Datacenter>
						</datacenter>
					</Datacenter>
				</DatacenterStruct>
			</ArrayOfDatacenterStruct>`,
			wantErr: false,
		},
		{
			name:    "invalid XML",
			xml:     "<invalid>",
			wantErr: true,
		},
		{
			name:    "empty XML",
			xml:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.xml)
			got, err := ParseDatacenterXML(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDatacenterXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("ParseDatacenterXML() returned nil result for valid XML")
			}
		})
	}
}

func TestParseStatusXML(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		want    *Status
		wantErr bool
	}{
		{
			name: "valid status XML",
			xml: `<Status>
				<logintierlastnumbers>100</logintierlastnumbers>
				<logintiers>tier1</logintiers>
				<queuenames>queue1</queuenames>
				<allow_billing_role>true</allow_billing_role>
				<queueurls>url1</queueurls>
				<lastassignedqueuenumber>50</lastassignedqueuenumber>
				<name>TestServer</name>
				<farmid>123</farmid>
				<world_full>false</world_full>
				<wait_hint>10</wait_hint>
			</Status>`,
			wantErr: false,
		},
		{
			name:    "invalid XML",
			xml:     "<invalid>",
			wantErr: true,
		},
		{
			name:    "empty XML",
			xml:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.xml)
			got, err := ParseStatusXML(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStatusXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Error("ParseStatusXML() returned nil result for valid XML")
				} else if got.Name != "TestServer" && tt.name == "valid status XML" {
					t.Errorf("ParseStatusXML() got Name = %v, want %v", got.Name, "TestServer")
				}
			}
		})
	}
}
