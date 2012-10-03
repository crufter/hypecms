package top

import(
	"github.com/opesun/hypecms/frame/misc/scut"
	"path/filepath"
	"strings"
	"net/http"
)

// Since we don't include the template name into the url, only "template", we have to extract the template name from the opt here.
// Example: xyz.com/template/style.css
//			xyz.com/tpl/admin/style.css
func (t *Top) serveFile() {
	uni := t.uni
	first_p := uni.Paths[1]
	last_p := uni.Paths[len(uni.Paths)-1]
	has_sfx := strings.HasSuffix(last_p, ".go")
	if first_p == "template" || first_p == "tpl" && !has_sfx {
		t.serveTemplateFile()
	} else if !has_sfx {
		if uni.Paths[1] == "shared" {
			http.ServeFile(uni.W, uni.Req, filepath.Join(t.config.AbsPath, uni.Req.URL.Path))
		} else {
			http.ServeFile(uni.W, uni.Req, filepath.Join(t.config.AbsPath, "uploads", uni.Req.Host, uni.Req.URL.Path))
		}
	} else {
		uni.Put("Don't do that.")
	}
}

func (t *Top) serveTemplateFile() {
	uni := t.uni
	if uni.Paths[1] == "template" {
		p := scut.GetTPath(uni.Opt, uni.Req.Host)
		http.ServeFile(uni.W, uni.Req, filepath.Join(uni.Root, p, strings.Join(uni.Paths[2:], "/")))
	} else { // "tpl"
		http.ServeFile(uni.W, uni.Req, filepath.Join(uni.Root, "modules", uni.Paths[2], "tpl", strings.Join(uni.Paths[3:], "/")))
	}
}