package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAlertContactsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "uptimerobot_alert_contacts" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.uptimerobot_alert_contacts.test", "alert_contacts.#", "1"),
					resource.TestCheckResourceAttr("data.uptimerobot_alert_contacts.test", "alert_contacts.0.id", "1"),
					resource.TestCheckResourceAttr("data.uptimerobot_alert_contacts.test", "alert_contacts.0.friendly_name", "foo"),
					resource.TestCheckResourceAttr("data.uptimerobot_alert_contacts.test", "alert_contacts.0.type", "email"),
					resource.TestCheckResourceAttr("data.uptimerobot_alert_contacts.test", "alert_contacts.0.status", "1"),
					resource.TestCheckResourceAttr("data.uptimerobot_alert_contacts.test", "alert_contacts.0.value", "foo@example.com"),
				),
			},
		},
	})
}
