# The example below is for a simple import that can be used when you do not have an inputs attribute block defined with secrets.
terraform import aap_credential.example 5

# If you have an inputs attribute object block defined with secrets, you need to specify them in the import command ID.
# The example below shows the pattern for the import command when you have an inputs attribute block defined with secrets in your .tf file.
# The ID field for the import command is the resources's ID followed by a comma-separated list of key/value pairs.
# Non-secret inputs do not need to be included in the import command
# The string at the end of this example command below would correlate to the following resource definition:
#   resource "aap_credential" "example_with_input" {
#      id = 5   
#      inputs = {
#         password = "12345"
#         token = "a1b2c3-d4e5-example"
#         non-secret = "do not include in import cli command"
#      }
#   }

terraform import aap_credential.example_with_input "5,password,12345,token,a1b2c3-d4e5-example"
