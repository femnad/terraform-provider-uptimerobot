package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccMonitorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "uptimerobot_monitor" "test" {
  friendly_name = "test"
  url = "http://example.com"
  type = "http"
  interval = 23
  timeout = 37
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "friendly_name", "test"),
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "url", "http://example.com"),
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "type", "http"),
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "interval", "23"),
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "timeout", "37"),
					resource.TestCheckResourceAttrSet("uptimerobot_monitor.test", "id"),
					resource.TestCheckResourceAttrSet("uptimerobot_monitor.test", "last_updated"),
				),
			},
			{
				ResourceName:            "uptimerobot_monitor.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			{
				Config: providerConfig + `
resource "uptimerobot_monitor" "test" {
  friendly_name = "test1"
  url = "http://example.com/foo"
  type = "http"
  interval = 27
  timeout = 34
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "friendly_name", "test1"),
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "url", "http://example.com/foo"),
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "type", "http"),
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "interval", "27"),
					resource.TestCheckResourceAttr("uptimerobot_monitor.test", "timeout", "34"),
				),
			},
		},
	})
}
