# terraform-provider-uptimerobot

A quick and dirty provider for interacting with [UptimeRobot](https://uptimerobot.com/) API. It doesn't implement all the features and **it's not well tested** but [documentation](https://registry.terraform.io/providers/femnad/uptimerobot/latest/docs) should be available for implemented resources.

## Usage

Basic usage example, demonstrating adding an HTTP monitor for `example.com`:

```tf
terraform {
  required_version = ">= 0.13"

  required_providers {
    uptimerobot = {
      source  = "femnad/uptimerobot"
      version = "0.1.0"
    }
  }
}

provider "uptimerobot" {
  # Alternative to setting API key via UPTIMEROBOT_API_KEY env var
  api_key = "<api-key>"
}

data "uptimerobot_account_details" "this" {
}

data "uptimerobot_alert_contact" "default" {
  type  = "e-mail"
  value = data.uptimerobot_account_details.this.email
}

resource "uptimerobot_monitor" "mta-sts" {
  friendly_name = "example"
  interval      = 3600
  type          = "http"
  url           = "http://example.com"

  alert_contact {
    id = data.uptimerobot_alert_contact.default.id
  }
}
```

## Alternatives

* [louy/terraform-provider-uptimerobot](https://github.com/louy/terraform-provider-uptimerobot)
