package main // import "github.com/go-gl/example/gl41core-cube"

type elem struct {
	left, bottom, width, height float32 
	depth float32 //this is used to do, well, depth sorting of background layers
	aspect_u float32 //what fraction of the element is "texture", width-wise [actually 1/frac, always > 1]
	atl_offu, atl_offv, atl_wid, atl_height float32 //offsets into the texture atlas + width/height of the subtextures
	tex_layer uint8 //this is really the selector bitfield, but the lowest bits of the selector are tex_layer
	texture	  uint8 //this is the next highest bits of the selector bitfield
	bits	uint16 //this is the remainder of the bitfield (used for various selectors)
	name string
	//signal handlers - the pane, paneatlas parsing setup process should register these appropriately with the paneatlas event handler
	value uint8 //an input bitfield type, indicating a data item that the texture uvs_atlas and/or selector values should be altered by
	action uint8 //an output type, indicating an action which is bound to clicking this button
}

type layout struct {
	elems []elem //the elements of the layout
	id	uint8 //the id
	name	string //and human-readable name of the layout
}

//proforma for layout elem

//			{
//				//x (from l), y (from bottom), width,height
//				//depth
//				//fraction of width with texture
//				//uv_atlas l,b,width,height
//				//tex layer
//				//texture unit
//				//bits
//				//name
//				//value binding (0=none)
//				//action binding (0=none)
//			},


//slice of layouts (each of which is a set of elems + id and name
var layouts = []layout{
	//MANAGER PANE
	layout{ 
		elems: []elem{
			{
				0.1,0.9,0.8,0.05,
				-1.0, //standard depth
				1.0, //fullwidth texture
				//0.0,0.0,0.1,0.2, //more like what we'd actually use if not debugging
				0.0,0.0,1.0,1.0, //a subtexture starting at 0,0, 100% width of atlas, 100% height of atlas
				0, //tex layer 0 (the "red" layer)
				0, //texture unit 0
				0, //bits all set to 0, so palette 0, not selected, not active
				"Title", //name of the element
				0,
				1,
			},
			{
				0.05, 0.05, 0.9, 0.8,//x (from l), x (from bottom), width,height
				-1.0,//depth
				0.8, //fraction of width with texture
				0.0,0.0,1.0,0.5,//uv_atlas l,b,width,height
				2,//tex layer
				0,//texture unit
				0,//bits
				"Example",//name
				0,//value binding (0=none)
				1,//action binding (0=none)
			},
		       },
		id: 0,
		name: "MANAGER",
		},
	//SCOREBOARD PANE
	layout{
		elems: []elem{
			{	//the background to TEAM1 half of display, which exists just to be a single colour rect
				0.0, 0.31, 0.5, 0.69, //one half of pane, as a column
				0.0, //background depth
				1.0, //fullwidth texture
				0.0, 0.0, 0.1, 0.1, //a zero sized subtexture which will definitely evaluate to "black"/"background"
				0, 
				0,
				0, //should actually set this to TEAMCOLOUR1 from palette
				"T1B",
				0,
				0,
			},
			{	//the background to TEAM2 half of display, which exists just to be a single colour rect
				0.5, 0.31, 0.5, 0.69, //one half of pane, as a column
				0.0, //background depth
				1.0, //fullwidth texture
				0.9, 0.9, 0.1, 0.1, //a zero sized subtexture which will definitely evaluate to "black"/"background"
				0, 
				0,
				0, //should actually set this to TEAMCOLOUR2 from palette
				"T2B",
				0,
				0,
			},
			{
				0.05, 0.82, 0.4, 0.15,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"TEAM1",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.55, 0.82, 0.4, 0.15,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"TEAM2",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.05, 0.44, 0.4, 0.35,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"SCORE1",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.55, 0.44, 0.4, 0.35,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"SCORE2",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.07, 0.32, 0.15, 0.1,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"PASS1",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.57, 0.32, 0.15, 0.1,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"PASS2",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.27, 0.32, 0.15, 0.1,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"PEN1",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.77, 0.32, 0.15, 0.1,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"PEN2",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.05, 0.05, 0.25, 0.20,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"PERIOD",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.32, 0.03, 0.36, 0.27,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"TIME",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},
			{
				0.70, 0.05, 0.25, 0.20,//x,y,width,height
				-1.0, //depth
				1.0,//fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				0,//tex layer
				0,//texture unit
				0,//bits
				"JAM",//name
				1,//input binding (0=none)
				0,//output binding (0=none)
			},

		       },
		id: 1,
		name: "SCOREBOARD",
		},
	//TIMEKEEPER PANE [time display + T/O options, STOP JAM option for early stop]
	layout{ elems: []elem{
			{
				0.05, 0.35, 0.9, 0.64,//x (from l), y (from bottom), width,height
				-1.0, //depth
				1.0, //fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				2,//tex layer
				0,//texture unit
				0,//bits
				"TIME",//name
				1,//value binding (0=none)
				1,//action binding (0=none)
			},
			{
				0.05, 0.05, 0.9, 0.29,//x (from l), y (from bottom), width,height
				-1.0, //depth
				1.0, //fraction of width with texture
				0.0,0.0,1.0,1.0,//uv_atlas l,b,width,height
				2,//tex layer
				0,//texture unit
				0,//bits
				"PAUSE",//name
				0,//value binding (0=none)
				1,//action binding (0=none)
			},
		       },
		id: 2,
		name: "TIMEKEEPER",
		},
	//SCOREKEEPER PANE [basic - pass + penalty buttons for each team]
	layout{
		elems: []elem{
			{	//the TEAM1 PASS
				0.01, 0.41, 0.48, 0.58, //one half of pane, as a column
				-1.0, //foreground
				1.0, //fullwidth texture
				0.0, 0.0, 0.1, 0.1, //a zero sized subtexture which will definitely evaluate to "black"/"background"
				0, 
				0,
				0, //should actually set this to TEAMCOLOUR1 from palette
				"TEAM1PASS",
				1, //value
				1, //action
			},
			{	//the TEAM2 pass	
				0.51, 0.41, 0.48, 0.58, //one half of pane, as a column
				-1.0, //foreground
				1.0, //fullwidth texture
				0.9, 0.9, 0.1, 0.1, //a zero sized subtexture which will definitely evaluate to "black"/"background"
				0, 
				0,
				0, //should actually set this to TEAMCOLOUR2 from palette
				"TEAM2PASS",
				1,
				1,
			},
			{	//the TEAM1 PEN
				0.01, 0.01, 0.48, 0.39, //one half of pane, as a column
				-1.0, //foreground
				1.0, //fullwidth texture
				0.0, 0.0, 0.1, 0.1, //a zero sized subtexture which will definitely evaluate to "black"/"background"
				0, 
				0,
				0, //should actually set this to TEAMCOLOUR1 from palette
				"TEAM1PEN",
				1,
				1,
			},
			{	//the TEAM2 PEN	
				0.51, 0.01, 0.48, 0.39, //one half of pane, as a column
				-1.0, //foreground
				1.0, //fullwidth texture
				0.9, 0.9, 0.1, 0.1, //a zero sized subtexture which will definitely evaluate to "black"/"background"
				0, 
				0,
				0, //should actually set this to TEAMCOLOUR2 from palette
				"TEAM2PEN",
				1,
				1,
			},
		       },
		id:3,
		name: "SCOREKEEPER",
		},
	//SCOREKEEPER PANE [details - pass button + 1-7 penalty buttons for each team]
	//SCOREKEEPER PANE [dynamic - pass button + 3 penalty buttons for each team (decided by optional lineup tracker)
	//TEAM SETUP PANE [optional] text input, colour selection here for both team rosters, names
	//LINEKEEPER PANE [optional] select each jam lineup per team
	//PENALTYKEEPER PANE [optional, details] #1
	//PENALTYKEEPER PANE [optional, dynamic] #1
	//PENALTYKEEPER PANE [optional] #2 (modal result of selecting an option on #1)
	//REMOTE MANAGER PANE [optional]
	//STATS DISPLAY PANE [optional]
}

