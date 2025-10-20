package shaders

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// TODO: select either one or use both
func CompileShaderFromFile(path string, shaderType uint32) (uint32, error) {
	sourceBytes, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read shader file %q: %v", path, err)
	}

	source := string(sourceBytes)

	if !strings.HasSuffix(source, "\x00") {
		source += "\x00"
	}

	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	defer free()

	gl.ShaderSource(shader, 1, csources, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		logMsg := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(logMsg))

		gl.DeleteShader(shader)
		return 0, fmt.Errorf("failed to compile %s shader: %v", path, strings.TrimSpace(logMsg))
	}

	return shader, nil
}

func CompileShaderFromSource(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &logMsg[0])
		return 0, fmt.Errorf("failed to compile shader: %s", logMsg)
	}

	return shader, nil
}
