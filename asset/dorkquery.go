package asset

var DorkQueries = []string{
	"intitle:\"index of\" setting.php",
	"intext:\"user\" filetype:php intext:\"account\" inurl:/admin",
	"intitle:\"Index of /confidential\"",
	"inurl:\"/wp-content/debug.log\"",
	"intext:\"index of\" \"infophp()\"",
	"intitle:\"Index of /confidential\"",
}
