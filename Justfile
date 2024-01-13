override := '''provider_installation {

  dev_overrides {
    "registry.terraform.io/femnad/uptimerobot" = "${GOPATH:-$HOME/go}/bin"
  }

  direct {}
}'''

add-override:
    echo "{{ override }}" > $HOME/.terraformrc

remove-override:
    rm -f $HOME/.terraformrc
