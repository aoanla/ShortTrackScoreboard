package main 


import (
	"fmt"
	"go/build"
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
//	"github.com/JamesMilnerUK/quadtree-go" //quadtrees for picking	
//	"github.com/go-gl/mathgl/mgl32"
)

//const windowWidth = 800
//const windowHeight = 600



func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	//it's probably easier to make set of 0..1 colliders for each pane, then to make a 0...0.5, 0.5...1 set?
	var p paneatlas

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	//tryout monitor stuff - OSX native fullscreen seems to break rendering (it's obviously not the same as other os fullscreens)
	monitor := glfw.GetPrimaryMonitor()
	vid_mode := monitor.GetVideoMode()
	windowWidth := vid_mode.Width
	windowHeight := vid_mode.Height

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	p.init(program)

	gl.UseProgram(program)

	width, height := window.GetFramebufferSize();
	gl.Viewport(0, 0, int32(width), int32(height) );

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	//load textures
	_, err = newTexture("output.png")
	tex_zero := gl.GetUniformLocation(program, gl.Str("texture1\x00")) 
	gl.Uniform1i(tex_zero,0) //"texture" is bound to GL_TEXTURE0, so we associate unit0 to shader uniform input 0	

	colour_mat := [12]float32{
		1.0,1.0,1.0, //colour1
		0.0,1.0,1.0, //colour2
		1.0,0.0,0.0, //colour3
		0.0,0.0,1.0, //colour4
	} 
	//associate the next uniform with our colour matrix
	//   this will be in location 1, as the texture is in 0 - until we add more textures, then this will move down?
	matrix_loc := gl.GetUniformLocation(program, gl.Str("colour_mat\x00"))
	//			loc, number of matrices, transpose?, pointer to matrix
	gl.UniformMatrix4x3fv(matrix_loc,1, false, &colour_mat[0])
	//this above will need rebinding, I think, if we change any of the colour values

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	previousTime := glfw.GetTime()

	//this needs to work on passing layout lists to p (which is now a paneatlas) not the old p (which was a pane)
	//boxes := make([]*box, 20)
		//selector 1 = g component, selector 2 = b component, hopefully	
		//p.newElem(0.0,0.0,0.2,0.2, 0.0,0.0,1.0,1.0, 1 ,"bottom")
		//p.newElem(0.25,0.25,0.7,0.4, 0.0,0.0, 0.5,0.5, 2 ,"top")
	p.loadLayouts(layouts) //layouts is a generic layout list

	p.updateVAO()
	//glfw.SetSwapInterval(1)

	// running in immediate mode, with checks every frame (and render every frame)
	// uses: ~4% CPU idle -> 7% CPU (max load, changing the selected box several times a second)

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		//f*ing high-def displays - width, height here do *not* map to the window pixel coords returned by mouse coords
		width, height = window.GetFramebufferSize();
		ww, wh := window.GetSize();
		// Update
		time := glfw.GetTime()
		if (time - previousTime > 4.0 ) {
			previousTime = time
		}
		xpos, ypos := window.GetCursorPos()
			//fmt.Printf("Mouse position: %f by %f\n", xpos, ypos)
			//fmt.Printf("Relative positions: %f by %f\n", xpos/float64(ww), 1.0 - ypos/float64(wh))
			//need to shift the positions to the coordinates of the quads (bl centred, not tl like window coords)
		p.mouse_select(xpos/float64(ww), 1.0 - ypos/float64(wh))


		// Render - this is now the paneatlas level render, which calls the renders for each subpane in sequence with the right viewport
		//this probably means that the paneatlas needs a window reference
		p.render(width, height)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewNRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

var vertexShader = `
#version 330

in vec3 vert;
in vec3 vertTexCoord; //z is the "wide" version of x which is 0..1 [z scales by the centering needed for texture coord]
in vec4 vertAtlasCoord;
in uint selector_i;
//when we're using a texture atlas, we use the below to specify the subset
// such that xy = origin, wz = width/height
//in vec4 vertTexCoordOffsets;
//this lets us do repeat / clamps of textures within the atlas (by clamping rangs on vertTexCoord)
// so that vTC < 0 > 1 map to 0, 1, and thus if we have our uv start at, say -1 ... 2 only the middle 0..1 will get texture

//also need an in for "selected" indicator, which is probably an int or a colour
//in vec4 colour;
//

out vec3 fragTexCoord;
out vec4 fragAtlasCoord;
flat out uint frag_selector;

void main() {
    fragTexCoord = vertTexCoord;
    fragAtlasCoord = vertAtlasCoord;
    frag_selector = selector_i;
    gl_Position = vec4(vert,1);
}
` + "\x00"

//rounded rectangle shader, used for items which need no contents
var fragmentShader = `
#version 330

uniform sampler2D texture1;
uniform mat4x3 colour_mat;
in vec3 fragTexCoord;
in vec4 fragAtlasCoord;
flat in uint frag_selector;
//in vec4 colour; //more of a tint really

out vec4 outputColor;

void main() {
	//rounded edges (radius is fraction of total width/length to round off)
	// this assumes max u,v are 0..1 so it will break with our thing for centering texture elements
	float u_radius = 0.05;
	if ( length(max(abs(fragTexCoord.xy-vec2(0.5))-0.5+u_radius, 0.0)) > u_radius) {
        	discard;
    	}
	vec2 samplecoord = fragAtlasCoord.xy + fragAtlasCoord.ba * vec2(clamp(fragTexCoord.z,0.0,1.0), fragTexCoord.y);
	float tmp = texture(texture1, samplecoord)[frag_selector & 3u];
	uint premul = (frag_selector & 32u) / 32u; //"highlight" selector
	outputColor = vec4(tmp*colour_mat[1]*(1u+premul) , 1);
	//outputColor.b = 0.0;
	//outputColor.r = 0.0; fragTexCoord.s;
	//outputColor.g = 1.0; fragTexCoord.t;
}
` + "\x00"

//non-rounded rectangle, with border, used for panes themselves
var fragmentShaderOL =  `
#version 330


in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
	//discard almost all fragments, as we only want to render the border (outer 1%)
        if ( fragTexCoord.x > 0.01 || fragTexCoord.x < 0.99 || fragTexCoord.y > 0.01 || fragTexCoord < 0.99 ) {
                discard;
        }

        outputColor = vec4(0.8,0.8,0.8,1.0)
}

` + "\x00"

//rounded rectangle with texture, used for text-containing elements, which is most of them
var fragmentShaderTex = `
#version 330

uniform sampler2D tex;

in vec3 fragTexCoord;
in vec4	fragAtlasCoord;

out vec4 outputColor;

void main() {
        //rounded edges (radius is fraction of total width/length to round off)
        float u_radius = 0.05;
        if ( length(max(abs(fragTexCoord-vec2(0.5))-0.5+u_radius, 0.0)) > u_radius) {
                discard;
        }

    outputColor = texture(tex, fragTexCoord);
}

` + "\x00" 

//text rendering is handled with its own internal shaders, which are not exposed to the general GUI interface

// Set the working directory to the root of Go package, so that its assets can be accessed.
func init() {
	//dir, err := importPathToDir("github.com/aoanla/ShortTrackScoreboard")
	//if err != nil {
	//	log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	//}
	err := os.Chdir(".")
	if err != nil {
		log.Panicln("os.Chdir:", err)
	}
}

// importPathToDir resolves the absolute path from importPath.
// There doesn't need to be a valid Go package inside that import path,
// but the directory must exist.
func importPathToDir(importPath string) (string, error) {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}
