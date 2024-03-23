# Manage example monitor.
resource "uptimerobot_monitor" "example" {
  friendly_name = "example"
  url           = "http://example.com"
  type          = "http"
  interval      = 73
  timeout       = 44
  alert_contact {
    id = "<alert-contact-id>"
  }
}
