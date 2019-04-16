package main // import "github.com/go-gl/example/gl41core-cube"

import (
	"fmt"
	"math"
	"github.com/go-gl/gl/v4.1-core/gl"
	//"github.com/go-gl/glfw/v3.2/glfw"
//	"github.com/JamesMilnerUK/quadtree-go" //quadtrees for picking	
//	"github.com/go-gl/mathgl/mgl32"
)

type pane struct {
	//vao uint32
	//vbo_xy, vbo_uv_atlas, vbo_selectors uint32
	//program uint32
	//textures uint32
	vertices []float32 //vec3
	//uvs []float32 		//vec2
	uvs_atlas []float32	//vec4 [offset+width,height]
	selectors	[]uint32 //bitfield selectors (1 per vertex) for texture selection etc
	boxoffsets []uint32	//where each box starts in the arrays
	collider *Quadtree
	selected	int16 //the box which has been selected (has a mouse over it)
	name	string
}

type paneatlas struct {
	vao uint32
	vbo_xy, vbo_uv_atlas, vbo_selectors uint32 //the pane atlas holds these (panes offset into them for their own ranges)
	program uint32
	textures uint32
	panes []*pane //all the panes which we have registered
	active_panes [4]uint8 //indexes into panes slice
	pane_offsets []uint32 //offsets of each pane into the vertex lists
	num_verts []uint32 //lengths of each pane in the vertex lists [range: pane_offset:pane_offset+num_verts]
	selected int8 //the pane which is selected
	vertices []float32 //interleaved vec3 (xyz) and vec3 (u',v,u)
	uvs_atlas []float32 //vec4
	selectors []uint32
}

const (
	_  = iota//texture plane selector 1
	_  = iota//texture plane selector 2
	_  = iota//texture unit selector 1
	_ = iota//texture unit selector 2
	_ = iota //texture unit selector 3
	SELECTED uint32 = 1 << iota //box selected indicator
	ACTIVATED //box "clicked"/"activated" indicator
	COLOUR_SELECT1 //colour selection bits (select out of the colour_mat)
	COLOUR_SELECT2
)

//just the bitfield versions of the above
const (
	SELECTED_BIT uint16 = 1 << iota
	ACTIVATED_BIT
	COLOUR_SELECT1_BIT
	COLOUR_SELECT2_BIT
)

// We organise our interface into panes, which handle stuff passed to them. The pane manager sets a Viewport and calls renderPane to render its elements
// Input events are picked up by the pane manager, and passed to the relevant pane (with their coordinates, for mouse etc, scaled to pane-relative values)
//
//func (pane) updatePane() {
//	//do any texture updates we need to do
//}
//

//what does a Layout look like? 
// - needs to have:
//		(coords, dims) (initial uv, uv_atlas, selector) for each box
//			event handlers				for each box
//			plumbing to attach to state
// glfw events [mouse, key, text]
// underlying data events [time, score, pass, penalties, lineup] - possibly handled by putting stuff in a channel, and triggering a glfx "empty" event to get everything to read its updates?

//the idea here is that the paneatlas pre-loads all the panes in one go, and assigns each of them a (static-length) section of the vertex buffer objects [and their CPU-size representations in the verticies etc arrays]
// so, the p's all have *slices* of the over-slice which the PA has (and the PA, only, has the power to commit the uberslice to the VBOs)
func (PA *paneatlas) loadLayouts(layouts []layout) (err int8) {
	var offsets uint32 = 0
	for _, item := range layouts {
		PA.pane_offsets = append(PA.pane_offsets, offsets) //starting offset for pane
		//have to do the slice thing backwards
		var p pane 
		p.init() //important initing step
		PA.panes = append(PA.panes, &p)
		p.parseLayout(item)
		//does the pane need to know its own offsets? Maybe only the PA needs to know this, as it's the one doing VBO changes
		//p.offset = offsets 
		//p.num_verts =  len(p.selectors)
		num_verts := uint32(len(p.selectors))
		PA.num_verts = append(PA.num_verts, num_verts)
		PA.vertices = append(PA.vertices, p.vertices...) //have to do these instead to make things work properly
		//and also for uvs_atlas, selectors
		PA.uvs_atlas = append(PA.uvs_atlas, p.uvs_atlas...) //have to do these instead to make things work properly
		PA.selectors = append(PA.selectors, p.selectors...) //have to do these instead to make things work properly
		//*now* all the underlying data matches up! (as long as no pane tries an append, in which case its slice will diverge from PAs)
		offsets += num_verts
	}
	//and then resync our p slices at the end (otherwise all but the last slice is invalidated by the append stuff making new slices above)
	for i, offsets := range PA.pane_offsets {
		p := PA.panes[i]
		num_verts := PA.num_verts[i]
		p.vertices = PA.vertices[offsets*6:(offsets+num_verts)*6]
		p.uvs_atlas = PA.uvs_atlas[offsets*4:(offsets+num_verts)*4] //uvs_atlas has 4 values per vertex
                p.selectors = PA.selectors[offsets:offsets+num_verts]
	}
//	PA.total_verts = offsets
	PA.updateVAO()
	fmt.Printf("Loaded: %d layouts\n", len(layouts))
	return 0 //think about errors next
}

func (P *pane) parseLayout(layout_ layout) (err int8) {
	for _, item := range layout_.elems {
		P.newElem(item)

	}
	P.name = layout_.name
	//also set pane name from layout_.name field
	return 0 //need to think about error handling
}
//		//get texture type
//		tex := texturetypein item
//		boxes[tex].append(new box)
//		//make the box vertices + push to the list too
//		vertices[tex].append(all the vertices)
//		// and the same for the uv and colours [colours default to 1,1,1]
//		//and push to the lists
//		//boxptr := newElem(item.left, item.bottom, item.width, item.height, programs[item.prog])
//		//register the uv coordinates of the box with whatever handles our event callbacks, if it's something which needs textures changed to reflect info.
//		//or we could do this dynamically, if we have an array with the uv coords which represent a given datum, which is updated when that datum changes
//		//and then we build the uv buffer each time, if it's dirtied
//		//note: always use glBufferSubData to update the dynamic uv buffers (it's always faster than BufferData, since we're not changing the size of the buffer, and it is much faster if only updating a small number of values
//	}
//}

func (P paneatlas) mouse_select(x_frac, y_frac float64) {
	//select into the right pane [if we were in 1 pane mode, we'd just pass the unaltered x_frac, y_frac to active_pane[0]]
	var pane_ uint8 = uint8(x_frac*2)&1 + (uint8(y_frac*2)&1)<<1 //width quadrant +1, height quadrant +2	
	var p = P.active_panes[pane_]
	//check that the pane we're in is the same pane as last time (and if it isn't, clear the previous pane's stuff)
	// TODO ********

	//regardless, collide into the currently active pane
	var offset1, offset2 = P.panes[p].mouse_select(2*math.Mod(x_frac,0.5), 2*math.Mod(y_frac, 0.5)) //rescale to pane coords
	//push GL state changes if we need to
	if (offset1 > -1 || offset2 > -1) {
		gl.BindVertexArray(P.vao)
                gl.BindBuffer(gl.ARRAY_BUFFER, P.vbo_selectors)
		if (offset1 > -1) {
			offset := uint32(offset1)*6+P.pane_offsets[p]
			gl.BufferSubData(gl.ARRAY_BUFFER, 4*int(offset), 6*4, gl.Ptr(P.selectors[offset:]))
		}
		if (offset2 > -1) {
			offset := uint32(offset2)*6+P.pane_offsets[p]
			gl.BufferSubData(gl.ARRAY_BUFFER, 4*int(offset), 6*4, gl.Ptr(P.selectors[offset:]))
		}
	}
	//
}


//pane is passed fractional coordinates in *its* coordinate system, scaled to its bounds
//This implementation is not proof against intersections with multiple overlapping elements 
//(but we also should never have 2 elements with actions overlapping)
func (P *pane) mouse_select(x_frac, y_frac float64) (selected, id int16){
	id = -1
	selected = P.selected
	intersections := P.collider.RetrieveIntersections(Bounds{
                                                X:      x_frac,
                                                Y:      y_frac,
                                                Width:  0,
                                                Height: 0,
                                                })
	if (len(intersections) > 0 ) { 
		id = intersections[0].id
	}
	//with pane atlas handler, we should only do selector updates, and pass update to parent pane atlas to tell it what subdata to push
	// the pane atlas consumes all the subdata offsets in the channel, and pushes each as BufferSubData updates from the master
	if (P.selected != id) { //box selected has changed, so do some stuff
		//this should be mutexed for gl operations, if we can do that?
		//push the selector buffer updates
		if (P.selected > -1){
			newval := P.selectors[P.selected*6] ^ SELECTED //these are all uniform values so we can assume first vertex is same for others
			for x :=(P.selected*6); x<(P.selected+1)*6 ; x++{
				P.selectors[x] = newval
			}
		}
		if (id > -1) {
			newval := P.selectors[id*6] ^ SELECTED //these are all uniform values so we can assume first vertex is same for others
			for x :=(id*6); x<(id+1)*6 ; x++{
				P.selectors[x] = newval
			}
		}
		P.selected = id //swap values at the end
		return
	}
	//if we get here, then there's no need to push GL state updates, so we explicitly -1 both offsets
	id = -1
	selected = -1 
	return
}

// stub function to activate the scoreboard window (which needs the context from the main window to share vertex data with it
//func (P *paneatlas) ActivateScoreboardWindow(primary_window *glfw.window) {
//	monitor := glfw.GetMonitors()[1] //I assume we'll want the second monitor
//	vidmode := monitor.GetVideoMode() 
//	P.scoreboard_ctx = glfw.CreateWindow(vidmode.Width, vidmode.Height, monitor, primary_window)
//	P.other_ctx = primary_window
//}


func (P paneatlas) render(width, height int){
	//this is for if we have a 4 pane view - a one pane view would use active_panes[0] only and the full width, height for it
	halfWidth := int32(width/2)
	halfHeight := int32(height/2)
	gl.UseProgram(P.program)
	gl.BindVertexArray(P.vao)
	for n,p := range P.active_panes { //
					//P.offsets[p] is offset into selectors
		gl.Viewport(halfWidth*int32(n & 1), halfHeight*int32((n & 2)>>1), halfWidth, halfHeight) 
		gl.DrawArrays(gl.TRIANGLES, int32(P.pane_offsets[p]), int32(P.num_verts[p]))
	}
	//paneatlas version of this is gl.DrawArrays(gl.TRIANGLES, P.offset, P.num_verts)
	//if (scoreboard_display == true) {
	//	also set the active window for the scoreboard, and draw all of it
	//	scoreboard_ctx.MakeContextCurrent()
	//	s_width, s_height := scoreboard_ctx.GetFrameBufferSize()
	//	gl.Viewport(0,0,int32(s_width), int32(s_height) //I assume this works in window-relative coords
	//	gl.DrawArrays(gl.TRIANGLES,int32(P.pane_offsets[1]), int32(P.num_verts[1])) //1 is always the scoreboard pane
	//	other_ctx.MakeContextCurrent() //and bind back to the "main" context

}
//takes coords in 0..1 space, 0 at bottom left - program is the glsl program to be used [if it's the same one we always use we can just set these bindings...]
func (P *paneatlas) init(program uint32) { //pass in the shader program here?
	gl.GenVertexArrays(1, &P.vao)
	gl.GenBuffers(1, &P.vbo_xy)
	gl.GenBuffers(1, &P.vbo_uv_atlas)	
	gl.GenBuffers(1, &P.vbo_selectors)
	P.program = program
	P.active_panes = [4]uint8{0,1,2,3} //BL, BR, TL, TR 
}

func (P *pane) init() {
	P.collider = &Quadtree{
                        Bounds: Bounds{
                                X: 0.0,
                                Y: 0.0,
                                Width: 1.0,
                                Height: 1.0,
                                },
                        MaxObjects: 10,
                        MaxLevels: 8,
                        Level: 0,
                        Objects: make([]Bounds, 0),
                        Nodes: make([]Quadtree, 0),
        }

	//P.program = program
	P.selected = -1 //no selected entity
}

func (P paneatlas) updateVAO () {
	gl.BindVertexArray(P.vao)
	
	//this is interleaved xyz, u,v,u_unnormalised vertex data
        gl.BindBuffer(gl.ARRAY_BUFFER, P.vbo_xy)
        gl.BufferData(gl.ARRAY_BUFFER, len(P.vertices)*4, gl.Ptr(P.vertices), gl.STATIC_DRAW)

        vertAttrib := uint32(gl.GetAttribLocation(P.program, gl.Str("vert\x00")))
        gl.EnableVertexAttribArray(vertAttrib)
        gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))

	//uv data will not actually change (because we texture items from a texture atlas
        texCoordAttrib := uint32(gl.GetAttribLocation(P.program, gl.Str("vertTexCoord\x00")))
        gl.EnableVertexAttribArray(texCoordAttrib)
        gl.VertexAttribPointer(texCoordAttrib, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))

	//uv atlas offset data [actually select our subtexture properly]. These are "pseudouniform" per primitive
	//								(that is, we give all vertices in prim same values)
	gl.BindBuffer(gl.ARRAY_BUFFER, P.vbo_uv_atlas)
	gl.BufferData(gl.ARRAY_BUFFER, len(P.uvs_atlas)*4, gl.Ptr(P.uvs_atlas), gl.DYNAMIC_DRAW) //so we store this with DYNAMIC hint

        texatlasCoordAttrib := uint32(gl.GetAttribLocation(P.program, gl.Str("vertAtlasCoord\x00")))
        gl.EnableVertexAttribArray(texatlasCoordAttrib)
        gl.VertexAttribPointer(texatlasCoordAttrib, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))

	//int selector value
	gl.BindBuffer(gl.ARRAY_BUFFER, P.vbo_selectors)
	gl.BufferData(gl.ARRAY_BUFFER, len(P.selectors)*4, gl.Ptr(P.selectors), gl.DYNAMIC_DRAW)
	
	selectorAttrib := uint32(gl.GetAttribLocation(P.program, gl.Str("selector_i\x00")))
	gl.EnableVertexAttribArray(selectorAttrib)
	gl.VertexAttribIPointer(selectorAttrib, 1, gl.UNSIGNED_INT, 4, gl.PtrOffset(0))
}

//func (P *pane) newElem(left, bottom, width, height float32, atl_offu, atl_offv, atl_wid, atl_height float32, tex_layer uint32, name string) {
func (P *pane) newElem(item elem) {

	//add new collider - don't add a collider if the elem has no ACTION set (item.action == 0)
	if (item.action != 0) {
		P.collider.Insert(Bounds{
			X: float64(item.left),
			Y: float64(item.bottom),
			Width: float64(item.width),
			Height: float64(item.height),
			id: int16(len(P.boxoffsets)),
		})
	}	
	
	//generate coordinates in opengl space (-1..1 range, with -1 at the bottom left)
	ll, bb := (item.left - 0.5)*2.0, (item.bottom - 0.5)*2.0
	rr, tt := ll + 2.0*item.width, bb + 2.0 * item.height
	
	//generate "unscaled" texture u from the item's "fraction of width which has the texture" 
	umin := float32( -(item.aspect_u - 1)/2.0 )
	umax := float32((item.aspect_u +1 ) /2.0 )
	//our specific vertex data for this box
	var boxverts = []float32 {
        //  X, Y, Z, U(scaled to 0..1), V, U (free range for positioning textures in part of prim)
        // Front
        ll, bb, item.depth,  1.0, 0.0, umax,
        rr, bb, item.depth,  0.0, 0.0, umin,
        ll, tt, item.depth,  1.0, 1.0, umax,
        rr, bb, item.depth,  0.0, 0.0, umin,
        rr, tt, item.depth,  0.0, 1.0, umin,
        ll, tt, item.depth,  1.0, 1.0, umax,
	}

	//uv_atlas
	var boxUVatl = []float32 {
	item.atl_offu, item.atl_offv, item.atl_wid, item.atl_height,
	item.atl_offu, item.atl_offv, item.atl_wid, item.atl_height,
	item.atl_offu, item.atl_offv, item.atl_wid, item.atl_height,
	item.atl_offu, item.atl_offv, item.atl_wid, item.atl_height,
	item.atl_offu, item.atl_offv, item.atl_wid, item.atl_height,
	item.atl_offu, item.atl_offv, item.atl_wid, item.atl_height,
	}

	//texture layer selection - most textures we render are monochrome [we tint them in the frag shader]
	// so we can treat each texture as "really" being 4 textures [r,g,b,a] = [0,1,2,3]
	//      this has the practical effect of multiplying the minimum number of bound textures we can have by 4
	// 		(as the min for compatibility is 8, this gives us an ample 32 layers in total)
	//		(we only really need 4 layers for most of the operations, so this is v efficient)
	//		(the remainder are mostly needed for the font atlas stuff)
	//	so, we need 1 texture unit selector, and 1 texture layer selector, but these both fit into 1 int
	//	we're also going to have a selector to pick the colour tint (from a palette in a uniform)
	//							and if highlighting is on
	//		   TL  TEX  BITS ("selected","active", colour palette etc)
	// selector val = [01][234][5..21]
	var selector uint32 = uint32(item.tex_layer & 3) + (uint32(item.texture & 7)<<2) +  (uint32(item.bits & 65535) << 5)
	var selectors = []uint32 {
		selector,
		selector,
		selector,
		selector,
		selector,
		selector,
	}
	
	//var boxRelUV = []float32 { }
	//var boxSelectCol = []float32 { } //float?
	P.vertices = append(P.vertices, boxverts...)
	P.uvs_atlas = append(P.uvs_atlas, boxUVatl...)
	P.selectors = append(P.selectors, selectors...)
	P.boxoffsets = append(P.boxoffsets, uint32(len(P.boxoffsets)))
}
