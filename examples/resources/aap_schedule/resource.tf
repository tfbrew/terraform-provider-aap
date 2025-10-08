resource "aap_schedule" "example" {
  name                 = "Example Schedule"
  unified_job_template = 1
  rrule                = "DTSTART;TZID=America/Chicago:20250124T090000 RRULE:INTERVAL=1;FREQ=WEEKLY;BYDAY=TU"
}

resource "aap_schedule" "test" {
  name                 = "Example With Extra Data"
  description          = "Example With Extra Data with showing all possible accepted data types"
  rrule                = "DTSTART;TZID=UTC:20250301T120000 RRULE:FREQ=DAILY;INTERVAL=1"
  unified_job_template = aap_job_template.our_template.id
  enabled              = true
  verbosity            = 2
  extra_data = {
    edString = "value1"
    edInt    = 3
    edFloat  = 4.3
    edList   = tolist(["two", "three"]) # Lists must be wrapped in tolist([...]) function
  }
}