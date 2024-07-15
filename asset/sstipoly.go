package asset

var SSTIPoly map[string][]string

func init() {
	SSTIPoly = make(map[string][]string)
	//source from https://cheatsheet.hackmanit.de/template-injection-table/index.html
	SSTIPoly["angularjs"] = []string{
		`class="ng-binding">p "&gt;[[$1]]`, `&lt;%=1%&gt;@*#{1}`, `Error`,
	}
	SSTIPoly["blade"] = []string{
		`p ">[[$1]]`, `Unmodified`, `Error`,
	}
	SSTIPoly["chameleon"] = []string{
		`p ">[[{1}]]`, `Unmodified`, `Unmodified`,
	}
	SSTIPoly["cheetah3"] = []string{
		`p ">[[{1}]]`, `1@*#{1}`, `{`,
	}
	SSTIPoly["django"] = []string{
		`Error`, `Unmodified`, `/**/`,
	}
	SSTIPoly["dot"] = []string{
		`Error`, `Unmodified`, `{##}`,
	}
	SSTIPoly["dotliquid"] = []string{
		`p ">[[$1]]`, `Unmodified`, `{##}/**/`,
	}
	SSTIPoly["eex"] = []string{
		`Unmodified`, `1@*#{1}`, `Unmodified`,
	}
	SSTIPoly["ejs"] = []string{
		`Unmodified`, `1@*#{1}`, `Unmodified`,
	}
	SSTIPoly["erb"] = []string{
		`Unmodified`, `1@*#{1}`, `Unmodified`,
	}
	SSTIPoly["erubi"] = []string{
		`Unmodified`, `1@*#{1}`, `Unmodified`,
	}
	SSTIPoly["erubis"] = []string{
		`Unmodified`, `1@*#{1}`, `Unmodified`,
	}
	SSTIPoly["eta"] = []string{
		`Unmodified`, `1@*#{1}`, `Unmodified`,
	}
	SSTIPoly["fluid"] = []string{
		`p ">[[$1]]`, `Unmodified`, `Error`,
	}
	SSTIPoly["freemarker"] = []string{
		`Error`, `<%=1%>@*1`, `Unmodified`,
	}
	SSTIPoly["groovy"] = []string{
		`p ">[[1]]`, `1@*#{1}`, `Unmodified`,
	}
	SSTIPoly["haml"] = []string{
		`Unmodified`, `<%=1%>@*1`, `Unmodified`,
	}
	SSTIPoly["handlebars"] = []string{
		`p ">[[$]]`, `Unmodified`, `{##}/**/`,
	}
	SSTIPoly["hoganjs"] = []string{
		`p ">[[$]]`, `Unmodified`, `{##}/**/`,
	}
	SSTIPoly["golang"] = []string{
		`p ">[[$1]]`, `&lt;%=1%>@*#{1}`, `{##}/*`,
	}
	SSTIPoly["jinja2"] = []string{
		`p ">[[$1]]`, `Unmodified`, `Error`,
	}
	SSTIPoly["latte"] = []string{
		`p ">[[${1}]]`, `<%=1%>@*#1`, `Error`,
	}
	SSTIPoly["liquid"] = []string{
		`p ">[[$1]]`, `Unmodified`, `{##}/**/`,
	}
	SSTIPoly["mako"] = []string{
		`p ">[[{1}]]`, `Error`, `Unmodified`,
	}
	SSTIPoly["mustache"] = []string{
		`p ">[[$]]`, `Unmodified`, `{##}/*`,
	}
	SSTIPoly["mustachejs"] = []string{
		`Error`, `Unmodified`, `{##}/**/`,
	}
	SSTIPoly["nunjucks"] = []string{
		`p ">[[$1]]`, `Unmodified`, `Error`,
	}
	SSTIPoly["pug"] = []string{
		`<p>">[[${{1}}]]</p>`, `<%=1%>@*1`, `Error`,
	}
	SSTIPoly["puginline"] = []string{
		`Unmodified`, `<%=1%>@*1`, `Unmodified`,
	}
	SSTIPoly["pystache"] = []string{
		`p ">[[$]]`, `Unmodified`, `{##}/**/`,
	}
	SSTIPoly["razorengine"] = []string{
		`Unmodified`, `<%=1%>`, `Unmodified`,
	}
	SSTIPoly["scriban"] = []string{
		`p ">[[$1]]`, `Unmodified`, `Error`,
	}
	SSTIPoly["slim"] = []string{
		`Unmodified`, `<%=1%>@*1`, `Unmodified`,
	}
	SSTIPoly["smarty"] = []string{
		`p ">[[$1]]`, `<%=1%>@*#1`, `Error`,
	}
	SSTIPoly["thymeleaf"] = []string{
		`p ">1`, `Unmodified`, `Unmodified`,
	}
	SSTIPoly["thymeleafinline"] = []string{
		`<a>p`, `Error`, `Error`,
	}
	SSTIPoly["tornado"] = []string{
		`p ">[[$1]]`, `Unmodified`, `Error`,
	}
	SSTIPoly["twig"] = []string{
		`p ">[[$1]]`, `Unmodified`, `Error`,
	}
	SSTIPoly["underscore"] = []string{
		`Unmodified`, `1@*#{1}`, `Unmodified`,
	}
	SSTIPoly["vuejs"] = []string{
		`p &quot;&gt;[[$1]]`, `&lt;%=1%&gt;@*#{1}`, `Error`,
	}
}
