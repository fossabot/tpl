package tpl

import (
	"context"
	"net/http"
	"strings"

	"code.gopub.tech/tpl/types"
)

// RenderToString 执行一个模板 并将结果输出为字符串
func RenderToString(tpl types.Template, data any) (string, error) {
	var sb strings.Builder
	err := tpl.Execute(&sb, data)
	return sb.String(), err
}

// NewHTMLRender 新建一个 HTML 渲染器
//
//	//go:embed views
//	var views embed.FS
//
//	var hotReload = gin.IsDebugging()
//	// NewHTMLRender
//	r, err := tpl.NewHTMLRender(func() (types.TemplateManager, error) {
//		m := html.NewTplManager()
//		if hotReload {
//			// 使用 os.DirFS 实时读取文件夹
//			return m, m.ParseWithSuffix(os.DirFS("views"), ".html")
//		}
//		// 使用编译时嵌入的 embed.FS 资源
//		return m, m.SetSubFS("views").ParseWithSuffix(views, ".html")
//	}, hotReload)
//
//	ginS.Get("/", func(c *gin.Context) {
//		// Instance
//		c.Render(http.StatusOK, r.Instance("index.html", gin.H{}))
//	})
//	ginS.Run()
func NewHTMLRender(builder types.Factory, opts ...NewHTMLRenderOpt) (types.ReloadableRender, error) {
	opt := &newHtmlRenderOpt{}
	for _, setter := range opts {
		setter(opt)
	}
	m, err := builder(context.Background())
	if err != nil {
		return nil, err
	}
	return &htmlRender{
		hotReload: opt.hotReload,
		builder:   builder,
		manager:   m,
	}, nil
}

type newHtmlRenderOpt struct {
	hotReload bool
}
type NewHTMLRenderOpt func(*newHtmlRenderOpt)

func WithHotReload(hotReload bool) NewHTMLRenderOpt {
	return func(o *newHtmlRenderOpt) { o.hotReload = hotReload }
}

type htmlRender struct {
	hotReload bool
	builder   types.Factory
	manager   types.TemplateManager
}

// Reload implements types.ReloadableRender.
func (h *htmlRender) Reload(ctx context.Context) error {
	m, err := h.builder(ctx)
	if err != nil {
		return nil
	}
	h.manager = m
	return nil
}

// Instance implements types.HTMLRender.
func (h *htmlRender) Instance(ctx context.Context, tplName string, data any) types.Render {
	tpl, err := h.GetTemplate(ctx, tplName)
	if err != nil {
		return &render{
			err: err,
		}
	}
	return &render{
		tpl: tpl,
		obj: data,
	}
}

// GetTemplate implements types.HTMLRender.
func (h *htmlRender) GetTemplate(ctx context.Context, tplName string) (types.Template, error) {
	var (
		m   = h.manager
		err error
	)
	if h.hotReload {
		m, err = h.builder(ctx)
	}
	if err != nil {
		return nil, err
	}
	return m.GetTemplate(tplName)
}

type render struct {
	err error
	tpl types.Template
	obj any
}

// Render implements types.Render.
func (r *render) Render(w http.ResponseWriter) error {
	if r.err != nil {
		return r.err
	}
	return r.tpl.Execute(w, r.obj)
}

var htmlContentType = []string{"text/html; charset=utf-8"}

// WriteContentType implements types.Render.
func (*render) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = htmlContentType
	}
}
