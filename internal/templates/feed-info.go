package templates

import "html/template"

// FeedInfoTemplate is the parsed template for /info command response
var FeedInfoTemplate *template.Template

// feedInfoTemplateHTML is the template for /info command response
const feedInfoTemplateHTML =
// language=GoTemplate
`📻 Podcast information

<b>{{.Title}}</b>

{{.Description}}
{{if .Author}}👨‍💻 By {{.Author}}{{end}}
🌐 Language: {{.Language}}
📚 Categories: {{categoriesList .Categories}}
{{if .Keywords}}🔑 Keywords: {{.Keywords}}{{end}}
{{if .ImageUrl}}🖼️ <a href="{{.ImageUrl}}">Artwork</a>{{end}}
{{if .WebsiteLink}}🔗 <a href="{{.WebsiteLink}}">Website</a>{{end}}
{{if gt .EpisodeCount 0}}🎧 Number of episodes: {{.EpisodeCount}}{{else}}📭 No episodes yet{{end}}
{{if .Explicit}}🔞 Explicit content{{end}}
{{if gt .EpisodeCount 0}}

📡 RSS: {{.RSSLink}}{{end}}`
