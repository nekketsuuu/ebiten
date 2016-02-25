// Copyright 2014 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// NOTICE: This file is not maintained well.

// +build android ios

package opengl

import (
	"errors"
	"fmt"

	mgl "golang.org/x/mobile/gl"
)

type Texture mgl.Texture
type Framebuffer mgl.Framebuffer
type Shader mgl.Shader
type Program mgl.Program
type Buffer mgl.Buffer

var ZeroFramebuffer Framebuffer

// TODO: Remove this after the GopherJS bug was fixed (#159)
func (p Program) Equals(other Program) bool {
	return p == other
}

type UniformLocation mgl.Uniform
type AttribLocation mgl.Attrib

type ProgramID uint32

func GetProgramID(p Program) ProgramID {
	return ProgramID(p.Value)
}

type context struct {
	worker mgl.Worker
	funcs  chan func()
}

// TODO: This variable can be in the context struct.
var (
	gl mgl.Context
)

func NewContext() *Context {
	c := &Context{
		Nearest:            mgl.NEAREST,
		Linear:             mgl.LINEAR,
		VertexShader:       mgl.VERTEX_SHADER,
		FragmentShader:     mgl.FRAGMENT_SHADER,
		ArrayBuffer:        mgl.ARRAY_BUFFER,
		ElementArrayBuffer: mgl.ELEMENT_ARRAY_BUFFER,
		DynamicDraw:        mgl.DYNAMIC_DRAW,
		StaticDraw:         mgl.STATIC_DRAW,
		Triangles:          mgl.TRIANGLES,
		Lines:              mgl.LINES,
	}
	c.funcs = make(chan func())
	gl, c.worker = mgl.NewContext()
	return c
}

func (c *Context) Loop() {
	for {
		select {
		case <-c.worker.WorkAvailable():
			c.worker.DoWork()
		case f := <-c.funcs:
			f()
		}
	}
}

func (c *Context) Init() {
	// This initialization must be done after Loop is called.
	// This is why Init is separated from NewContext.

	// Textures' pixel formats are alpha premultiplied.
	gl.Enable(mgl.BLEND)
	gl.BlendFunc(mgl.ONE, mgl.ONE_MINUS_SRC_ALPHA)
}

func (c *Context) RunOnContextThread(f func()) {
	ch := make(chan struct{})
	c.funcs <- func() {
		f()
		close(ch)
	}
	<-ch
	return
}

func (c *Context) Check() {
	if e := gl.GetError(); e != mgl.NO_ERROR {
		panic(fmt.Sprintf("check failed: %d", e))
	}
}

func (c *Context) NewTexture(width, height int, pixels []uint8, filter Filter) (Texture, error) {
	t := gl.CreateTexture()
	if t.Value <= 0 {
		return Texture{}, errors.New("opengl: creating texture failed")
	}
	gl.PixelStorei(mgl.UNPACK_ALIGNMENT, 4)
	gl.BindTexture(mgl.TEXTURE_2D, t)

	gl.TexParameteri(mgl.TEXTURE_2D, mgl.TEXTURE_MAG_FILTER, int(filter))
	gl.TexParameteri(mgl.TEXTURE_2D, mgl.TEXTURE_MIN_FILTER, int(filter))

	var p []uint8
	if pixels != nil {
		p = pixels
	}
	gl.TexImage2D(mgl.TEXTURE_2D, 0, width, height, mgl.RGBA, mgl.UNSIGNED_BYTE, p)

	return Texture(t), nil
}

func (c *Context) FramebufferPixels(f Framebuffer, width, height int) ([]uint8, error) {
	gl.Flush()

	gl.BindFramebuffer(mgl.FRAMEBUFFER, mgl.Framebuffer(f))

	pixels := make([]uint8, 4*width*height)
	gl.ReadPixels(pixels, 0, 0, width, height, mgl.RGBA, mgl.UNSIGNED_BYTE)
	if e := gl.GetError(); e != mgl.NO_ERROR {
		return nil, fmt.Errorf("opengl: glReadPixels: %d", e)
	}
	return pixels, nil
}

func (c *Context) BindTexture(t Texture) {
	gl.BindTexture(mgl.TEXTURE_2D, mgl.Texture(t))
}

func (c *Context) DeleteTexture(t Texture) {
	gl.DeleteTexture(mgl.Texture(t))
}

func (c *Context) TexSubImage2D(p []uint8, width, height int) {
	gl.TexSubImage2D(mgl.TEXTURE_2D, 0, 0, 0, width, height, mgl.RGBA, mgl.UNSIGNED_BYTE, p)
}

func (c *Context) BindZeroFramebuffer() {
	gl.BindFramebuffer(mgl.FRAMEBUFFER, mgl.Framebuffer(ZeroFramebuffer))
}

func (c *Context) NewFramebuffer(texture Texture) (Framebuffer, error) {
	f := gl.CreateFramebuffer()
	if f.Value <= 0 {
		return Framebuffer{}, errors.New("opengl: creating framebuffer failed: gl.IsFramebuffer returns false")
	}
	gl.BindFramebuffer(mgl.FRAMEBUFFER, f)

	gl.FramebufferTexture2D(mgl.FRAMEBUFFER, mgl.COLOR_ATTACHMENT0, mgl.TEXTURE_2D, mgl.Texture(texture), 0)
	s := gl.CheckFramebufferStatus(mgl.FRAMEBUFFER)
	if s != mgl.FRAMEBUFFER_COMPLETE {
		if s != 0 {
			return Framebuffer{}, fmt.Errorf("opengl: creating framebuffer failed: %v", s)
		}
		if e := gl.GetError(); e != mgl.NO_ERROR {
			return Framebuffer{}, fmt.Errorf("opengl: creating framebuffer failed: (glGetError) %d", e)
		}
		return Framebuffer{}, fmt.Errorf("opengl: creating framebuffer failed: unknown error")
	}

	return Framebuffer(f), nil
}

func (c *Context) SetViewport(f Framebuffer, width, height int) error {
	gl.Flush()
	gl.BindFramebuffer(mgl.FRAMEBUFFER, mgl.Framebuffer(f))
	if err := gl.CheckFramebufferStatus(mgl.FRAMEBUFFER); err != mgl.FRAMEBUFFER_COMPLETE {
		if e := gl.GetError(); e != 0 {
			return fmt.Errorf("opengl: glBindFramebuffer failed: %d", e)
		}
		return errors.New("opengl: glBindFramebuffer failed: the context is different?")
	}
	gl.Viewport(0, 0, width, height)
	return nil
}

func (c *Context) FillFramebuffer(r, g, b, a float64) error {
	gl.ClearColor(float32(r), float32(g), float32(b), float32(a))
	gl.Clear(mgl.COLOR_BUFFER_BIT)
	return nil
}

func (c *Context) DeleteFramebuffer(f Framebuffer) {
	gl.DeleteFramebuffer(mgl.Framebuffer(f))
}

func (c *Context) NewShader(shaderType ShaderType, source string) (Shader, error) {
	s := gl.CreateShader(mgl.Enum(shaderType))
	if s.Value == 0 {
		return Shader{}, errors.New("opengl: glCreateShader failed")
	}
	gl.ShaderSource(s, source)
	gl.CompileShader(s)

	v := gl.GetShaderi(s, mgl.COMPILE_STATUS)
	if v == mgl.FALSE {
		log := gl.GetShaderInfoLog(s)
		return Shader{}, fmt.Errorf("opengl: shader compile failed: %s", log)
	}
	return Shader(s), nil
}

func (c *Context) DeleteShader(s Shader) {
	gl.DeleteShader(mgl.Shader(s))
}

func (c *Context) GlslHighpSupported() bool {
	return false
}

func (c *Context) NewProgram(shaders []Shader) (Program, error) {
	p := gl.CreateProgram()
	if p.Value == 0 {
		return Program{}, errors.New("opengl: glCreateProgram failed")
	}

	for _, shader := range shaders {
		gl.AttachShader(p, mgl.Shader(shader))
	}
	gl.LinkProgram(p)
	v := gl.GetProgrami(p, mgl.LINK_STATUS)
	if v == mgl.FALSE {
		return Program{}, errors.New("opengl: program error")
	}
	return Program(p), nil
}

func (c *Context) UseProgram(p Program) {
	gl.UseProgram(mgl.Program(p))
}

func (c *Context) GetUniformLocation(p Program, location string) UniformLocation {
	u := UniformLocation(gl.GetUniformLocation(mgl.Program(p), location))
	if u.Value == -1 {
		panic("invalid uniform location: " + location)
	}
	return u
}

func (c *Context) UniformInt(p Program, location string, v int) {
	gl.Uniform1i(mgl.Uniform(GetUniformLocation(c, p, location)), v)
}

func (c *Context) UniformFloats(p Program, location string, v []float32) {
	l := mgl.Uniform(GetUniformLocation(c, p, location))
	switch len(v) {
	case 4:
		gl.Uniform4fv(l, v)
	case 16:
		gl.UniformMatrix4fv(l, v)
	default:
		panic("not reach")
	}
}

func (c *Context) GetAttribLocation(p Program, location string) AttribLocation {
	a := AttribLocation(gl.GetAttribLocation(mgl.Program(p), location))
	if a.Value == ^uint(0) {
		panic("invalid attrib location: " + location)
	}
	return a
}

func (c *Context) VertexAttribPointer(p Program, location string, normalize bool, stride int, size int, v int) {
	l := GetAttribLocation(c, p, location)
	gl.VertexAttribPointer(mgl.Attrib(l), size, mgl.SHORT, normalize, stride, v)
}

func (c *Context) EnableVertexAttribArray(p Program, location string) {
	l := GetAttribLocation(c, p, location)
	gl.EnableVertexAttribArray(mgl.Attrib(l))
}

func (c *Context) DisableVertexAttribArray(p Program, location string) {
	l := GetAttribLocation(c, p, location)
	gl.DisableVertexAttribArray(mgl.Attrib(l))
}

func uint16ToBytes(v []uint16) []byte {
	b := make([]byte, len(v)*2)
	for i, x := range v {
		b[2*i] = byte(x)
		b[2*i+1] = byte(x >> 8)
	}
	return b
}

func int16ToBytes(v []int16) []byte {
	b := make([]byte, len(v)*2)
	for i, x := range v {
		b[2*i] = byte(uint16(x))
		b[2*i+1] = byte(uint16(x) >> 8)
	}
	return b
}

func (c *Context) NewBuffer(bufferType BufferType, v interface{}, bufferUsage BufferUsage) Buffer {
	b := gl.CreateBuffer()
	gl.BindBuffer(mgl.Enum(bufferType), b)
	switch v := v.(type) {
	case int:
		gl.BufferInit(mgl.Enum(bufferType), v, mgl.Enum(bufferUsage))
		return Buffer(b)
	case []uint16:
		gl.BufferData(mgl.Enum(bufferType), uint16ToBytes(v), mgl.Enum(bufferUsage))
	default:
		panic("not reach")
	}
	return Buffer(b)
}

func (c *Context) BindElementArrayBuffer(b Buffer) {
	gl.BindBuffer(mgl.ELEMENT_ARRAY_BUFFER, mgl.Buffer(b))
}

func (c *Context) BufferSubData(bufferType BufferType, data []int16) {
	gl.BufferSubData(mgl.Enum(bufferType), 0, int16ToBytes(data))
}

func (c *Context) DrawElements(mode Mode, len int) {
	gl.DrawElements(mgl.Enum(mode), len, mgl.UNSIGNED_SHORT, 0)
}

func (c *Context) Finish() {
	gl.Finish()
}