resource "mailjet_sender" "sender_example" {
  email = "*@mailjet.example.com"
  name = "My mailjet sender example"
  is_default_sender = false
  email_type = "unknown"
}

# You can retrieve the DNS entries to set using the mailjet_dns data source

resource "mailjet_sender_validate" "sender_validate_example" {
    id = mailjet_sender.sender_example.id
}