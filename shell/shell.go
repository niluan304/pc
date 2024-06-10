package shell

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Shell
// 输入 shell 脚本文件，或者 shell 命令
func Shell(sh string) (out []byte, err error) {
	// 如果输入的是 shell脚本，将 shell脚本 读取出来
	// 如果有错误，最多文件不存在，意味着 text 为空，那就取原本的命令
	text, _ := os.ReadFile(sh)
	sh = cmp.Or(string(text), sh)

	// 创建命令对象
	cmd := exec.Command(shell, arg, sh)

	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("sh: %s", sh))
	}
	return out, nil
}
