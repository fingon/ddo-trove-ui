package templates

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html" //nolint:revive,staticcheck
)

func Layout(title string, children ...g.Node) g.Node {
	return Doctype(
		HTML(Lang("en"),
			Head(
				Meta(Charset("UTF-8")),
				Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
				TitleEl(g.Text(title)),
				Script(Src("https://unpkg.com/htmx.org@1.9.10"),
					Integrity("sha384-D1Kt99CQMDuVetoL1lrYwg5t+9QdHe7NLX/SoJYkXDFfX37iInKRy5xLSi8nO7UC"),
					g.Attr("crossorigin", "anonymous")),
				Link(Rel("stylesheet"), Href("/static/style.css")),
			),
			Body(
				Div(Class("container"),
					g.Group(children),
				),
			),
		),
	)
}
