package compose

// statsArgs streams the usage of every running Container, whatever Project it
// belongs to.
var statsArgs = []string{"stats", "--format", "json"}

// infoArgs asks the daemon for the size of the machine.
var infoArgs = []string{"info", "--format", "json"}
