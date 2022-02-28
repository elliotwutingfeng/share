// Code generated by github.com/shuLhan/share/lib/memfs DO NOT EDIT.

package embed

import (
	"github.com/shuLhan/share/lib/memfs"
)

func generate_testdata() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata",
		Path:            "/",
		ContentType:     "",
		GenFuncName:     "generate_testdata",
	}
	node.SetMode(2147484141)
	node.SetName("/")
	node.SetSize(0)
	node.AddChild(_memFS_getNode(memFS, "/direct", generate_testdata_direct))
	node.AddChild(_memFS_getNode(memFS, "/exclude", generate_testdata_exclude))
	node.AddChild(_memFS_getNode(memFS, "/include", generate_testdata_include))
	node.AddChild(_memFS_getNode(memFS, "/index.css", generate_testdata_index_css))
	node.AddChild(_memFS_getNode(memFS, "/index.html", generate_testdata_index_html))
	node.AddChild(_memFS_getNode(memFS, "/index.js", generate_testdata_index_js))
	node.AddChild(_memFS_getNode(memFS, "/plain", generate_testdata_plain))
	return node
}

func generate_testdata_direct() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/direct",
		Path:            "/direct",
		ContentType:     "",
		GenFuncName:     "generate_testdata_direct",
	}
	node.SetMode(2147484141)
	node.SetName("direct")
	node.SetSize(0)
	node.AddChild(_memFS_getNode(memFS, "/direct/add", generate_testdata_direct_add))
	return node
}

func generate_testdata_direct_add() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/direct/add",
		Path:            "/direct/add",
		ContentType:     "",
		GenFuncName:     "generate_testdata_direct_add",
	}
	node.SetMode(2147484141)
	node.SetName("add")
	node.SetSize(0)
	node.AddChild(_memFS_getNode(memFS, "/direct/add/file", generate_testdata_direct_add_file))
	node.AddChild(_memFS_getNode(memFS, "/direct/add/file2", generate_testdata_direct_add_file2))
	return node
}

func generate_testdata_direct_add_file() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/direct/add/file",
		Path:            "/direct/add/file",
		ContentType:     "text/plain; charset=utf-8",
		GenFuncName:     "generate_testdata_direct_add_file",
		Content:         []byte("\x54\x65\x73\x74\x20\x64\x69\x72\x65\x63\x74\x20\x61\x64\x64\x20\x66\x69\x6C\x65\x2E\x0A"),
	}
	node.SetMode(420)
	node.SetName("file")
	node.SetSize(22)
	return node
}

func generate_testdata_direct_add_file2() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/direct/add/file2",
		Path:            "/direct/add/file2",
		ContentType:     "text/plain; charset=utf-8",
		GenFuncName:     "generate_testdata_direct_add_file2",
		Content:         []byte("\x54\x65\x73\x74\x20\x64\x69\x72\x65\x63\x74\x20\x61\x64\x64\x20\x66\x69\x6C\x65\x20\x32\x2E\x0A"),
	}
	node.SetMode(420)
	node.SetName("file2")
	node.SetSize(24)
	return node
}

func generate_testdata_exclude() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/exclude",
		Path:            "/exclude",
		ContentType:     "",
		GenFuncName:     "generate_testdata_exclude",
	}
	node.SetMode(2147484141)
	node.SetName("exclude")
	node.SetSize(0)
	node.AddChild(_memFS_getNode(memFS, "/exclude/index-link.css", generate_testdata_exclude_index_link_css))
	node.AddChild(_memFS_getNode(memFS, "/exclude/index-link.html", generate_testdata_exclude_index_link_html))
	node.AddChild(_memFS_getNode(memFS, "/exclude/index-link.js", generate_testdata_exclude_index_link_js))
	return node
}

func generate_testdata_exclude_index_link_css() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/exclude/index-link.css",
		Path:            "/exclude/index-link.css",
		ContentType:     "text/css; charset=utf-8",
		GenFuncName:     "generate_testdata_exclude_index_link_css",
		Content:         []byte("\x62\x6F\x64\x79\x20\x7B\x0A\x7D\x0A"),
	}
	node.SetMode(420)
	node.SetName("index-link.css")
	node.SetSize(9)
	return node
}

func generate_testdata_exclude_index_link_html() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/exclude/index-link.html",
		Path:            "/exclude/index-link.html",
		ContentType:     "text/html; charset=utf-8",
		GenFuncName:     "generate_testdata_exclude_index_link_html",
		Content:         []byte("\x3C\x68\x74\x6D\x6C\x3E\x3C\x2F\x68\x74\x6D\x6C\x3E\x0A"),
	}
	node.SetMode(420)
	node.SetName("index-link.html")
	node.SetSize(14)
	return node
}

func generate_testdata_exclude_index_link_js() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/exclude/index-link.js",
		Path:            "/exclude/index-link.js",
		ContentType:     "text/javascript; charset=utf-8",
		GenFuncName:     "generate_testdata_exclude_index_link_js",
		Content:         []byte("\x66\x75\x6E\x63\x74\x69\x6F\x6E\x20\x58\x28\x29\x20\x7B\x7D\x0A"),
	}
	node.SetMode(420)
	node.SetName("index-link.js")
	node.SetSize(16)
	return node
}

func generate_testdata_include() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/include",
		Path:            "/include",
		ContentType:     "",
		GenFuncName:     "generate_testdata_include",
	}
	node.SetMode(2147484141)
	node.SetName("include")
	node.SetSize(0)
	node.AddChild(_memFS_getNode(memFS, "/include/index.css", generate_testdata_include_index_css))
	node.AddChild(_memFS_getNode(memFS, "/include/index.html", generate_testdata_include_index_html))
	node.AddChild(_memFS_getNode(memFS, "/include/index.js", generate_testdata_include_index_js))
	return node
}

func generate_testdata_include_index_css() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/include/index.css",
		Path:            "/include/index.css",
		ContentType:     "text/css; charset=utf-8",
		GenFuncName:     "generate_testdata_include_index_css",
		Content:         []byte("\x62\x6F\x64\x79\x20\x7B\x0A\x7D\x0A"),
	}
	node.SetMode(420)
	node.SetName("index.css")
	node.SetSize(9)
	return node
}

func generate_testdata_include_index_html() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/include/index.html",
		Path:            "/include/index.html",
		ContentType:     "text/html; charset=utf-8",
		GenFuncName:     "generate_testdata_include_index_html",
		Content:         []byte("\x3C\x68\x74\x6D\x6C\x3E\x3C\x2F\x68\x74\x6D\x6C\x3E\x0A"),
	}
	node.SetMode(420)
	node.SetName("index.html")
	node.SetSize(14)
	return node
}

func generate_testdata_include_index_js() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/include/index.js",
		Path:            "/include/index.js",
		ContentType:     "text/javascript; charset=utf-8",
		GenFuncName:     "generate_testdata_include_index_js",
		Content:         []byte("\x66\x75\x6E\x63\x74\x69\x6F\x6E\x20\x58\x28\x29\x20\x7B\x7D\x0A"),
	}
	node.SetMode(420)
	node.SetName("index.js")
	node.SetSize(16)
	return node
}

func generate_testdata_index_css() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/index.css",
		Path:            "/index.css",
		ContentType:     "text/css; charset=utf-8",
		GenFuncName:     "generate_testdata_index_css",
		Content:         []byte("\x62\x6F\x64\x79\x20\x7B\x0A\x7D\x0A"),
	}
	node.SetMode(420)
	node.SetName("index.css")
	node.SetSize(9)
	return node
}

func generate_testdata_index_html() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/index.html",
		Path:            "/index.html",
		ContentType:     "text/html; charset=utf-8",
		GenFuncName:     "generate_testdata_index_html",
		Content:         []byte("\x3C\x68\x74\x6D\x6C\x3E\x3C\x2F\x68\x74\x6D\x6C\x3E\x0A"),
	}
	node.SetMode(420)
	node.SetName("index.html")
	node.SetSize(14)
	return node
}

func generate_testdata_index_js() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/index.js",
		Path:            "/index.js",
		ContentType:     "text/javascript; charset=utf-8",
		GenFuncName:     "generate_testdata_index_js",
		Content:         []byte("\x66\x75\x6E\x63\x74\x69\x6F\x6E\x20\x58\x28\x29\x20\x7B\x7D\x0A"),
	}
	node.SetMode(420)
	node.SetName("index.js")
	node.SetSize(16)
	return node
}

func generate_testdata_plain() *memfs.Node {
	node := &memfs.Node{
		SysPath:         "testdata/plain",
		Path:            "/plain",
		ContentType:     "text/plain; charset=utf-8",
		GenFuncName:     "generate_testdata_plain",
		Content:         []byte("\x54\x68\x69\x73\x20\x69\x73\x20\x61\x20\x70\x6C\x61\x69\x6E\x20\x74\x65\x78\x74\x2E\x0A"),
	}
	node.SetMode(420)
	node.SetName("plain")
	node.SetSize(22)
	return node
}

//
// _memFS_getNode is internal function to minimize duplicate node
// created on Node.AddChild() and on generatedPathNode.Set().
//
func _memFS_getNode(mfs *memfs.MemFS, path string, fn func() *memfs.Node) (node *memfs.Node) {
	node = mfs.PathNodes.Get(path)
	if node != nil {
		return node
	}
	return fn()
}

func init() {
	memFS = &memfs.MemFS{
		PathNodes: memfs.NewPathNode(),
		Opts: &memfs.Options{
			Root: "testdata",
			MaxFileSize: 5242880,
			Includes: []string{
			},
			Excludes: []string{
				`^\..*`,
				`.*/node_save$`,
			},
			Embed: memfs.EmbedOptions{
				CommentHeader:  ``,
				PackageName:    "embed",
				VarName:        "memFS",
				GoFileName:     "./internal/test/embed_disable_modtime/embed_test.go",
				WithoutModTime: true,
			},
		},
	}
	memFS.PathNodes.Set("/",
		_memFS_getNode(memFS, "/", generate_testdata))
	memFS.PathNodes.Set("/direct",
		_memFS_getNode(memFS, "/direct", generate_testdata_direct))
	memFS.PathNodes.Set("/direct/add",
		_memFS_getNode(memFS, "/direct/add", generate_testdata_direct_add))
	memFS.PathNodes.Set("/direct/add/file",
		_memFS_getNode(memFS, "/direct/add/file", generate_testdata_direct_add_file))
	memFS.PathNodes.Set("/direct/add/file2",
		_memFS_getNode(memFS, "/direct/add/file2", generate_testdata_direct_add_file2))
	memFS.PathNodes.Set("/exclude",
		_memFS_getNode(memFS, "/exclude", generate_testdata_exclude))
	memFS.PathNodes.Set("/exclude/index-link.css",
		_memFS_getNode(memFS, "/exclude/index-link.css", generate_testdata_exclude_index_link_css))
	memFS.PathNodes.Set("/exclude/index-link.html",
		_memFS_getNode(memFS, "/exclude/index-link.html", generate_testdata_exclude_index_link_html))
	memFS.PathNodes.Set("/exclude/index-link.js",
		_memFS_getNode(memFS, "/exclude/index-link.js", generate_testdata_exclude_index_link_js))
	memFS.PathNodes.Set("/include",
		_memFS_getNode(memFS, "/include", generate_testdata_include))
	memFS.PathNodes.Set("/include/index.css",
		_memFS_getNode(memFS, "/include/index.css", generate_testdata_include_index_css))
	memFS.PathNodes.Set("/include/index.html",
		_memFS_getNode(memFS, "/include/index.html", generate_testdata_include_index_html))
	memFS.PathNodes.Set("/include/index.js",
		_memFS_getNode(memFS, "/include/index.js", generate_testdata_include_index_js))
	memFS.PathNodes.Set("/index.css",
		_memFS_getNode(memFS, "/index.css", generate_testdata_index_css))
	memFS.PathNodes.Set("/index.html",
		_memFS_getNode(memFS, "/index.html", generate_testdata_index_html))
	memFS.PathNodes.Set("/index.js",
		_memFS_getNode(memFS, "/index.js", generate_testdata_index_js))
	memFS.PathNodes.Set("/plain",
		_memFS_getNode(memFS, "/plain", generate_testdata_plain))

	memFS.Root = memFS.PathNodes.Get("/")
}
