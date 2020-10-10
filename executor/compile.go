package executor

import (
	"fmt"
	"github.com/LanceLRQ/deer-executor/provider"
	"path"
	"strings"
)

// 匹配编程语言
func matchCodeLanguage(keyword string, fileName string) (provider.CodeCompileProviderInterface, error) {
_match:
	switch keyword {
	case "c", "gcc", "gnu-c":
		return &provider.GnucCompileProvider{}, nil
	case "cpp", "gcc-cpp", "gpp", "g++":
		return &provider.GnucppCompileProvider{}, nil
	case "java":
		return &provider.JavaCompileProvider{}, nil
	case "py2", "python2":
		return &provider.Py2CompileProvider{}, nil
	case "py", "py3", "python3":
		return &provider.Py3CompileProvider{}, nil
	case "php":
		return &provider.PHPCompileProvider{}, nil
	case "node", "nodejs":
		return &provider.NodeJSCompileProvider{}, nil
	case "rb", "ruby":
		return &provider.RubyCompileProvider{}, nil
	case "auto":
		keyword = strings.Replace(path.Ext(fileName), ".", "", -1)
		goto _match
	}
	return nil, fmt.Errorf("unsupported language")
}

// 编译文件
// 如果不设置codeStr，默认会读取配置文件里的code_file字段并打开对应文件
func (options *JudgeOptions) getCompiler(codeStr string) (provider.CodeCompileProviderInterface, error) {
	if codeStr == "" {
		codeFileBytes, err := ReadFile(options.CodeFile)
		if err != nil {
			return nil, err
		}
		codeStr = string(codeFileBytes)
	}

	compiler, err := matchCodeLanguage(options.CodeLangName, options.CodeFile)
	if err != nil { return nil, err }
	err = compiler.Init(codeStr, "/tmp")
	if err != nil {
		return nil, err
	}
	return compiler, err
}