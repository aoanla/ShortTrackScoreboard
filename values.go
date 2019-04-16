package main

//reference for bitfield
// const (
//        _  = iota//texture plane selector 1
//        _  = iota//texture plane selector 2
//        _  = iota//texture unit selector 1
//        _ = iota//texture unit selector 2
//        _ = iota //texture unit selector 3
//        SELECTED uint32 = 1 << iota //box selected indicator
//        ACTIVATED //box "clicked"/"activated" indicator
//        COLOUR_SELECT1 //colour selection bits (select out of the colour_mat)
//        COLOUR_SELECT2
//)


//VALUEs for linked elem field


// - consider the possibility that we also need per-jam versions of Score and Penalty tags
const (
	VALUE_NULL uint16 = iota
	VALUE_SCORE_T1		//SCORES
	VALUE_SCORE_T2
	VALUE_PENALTY_T1	//PENALTIES - TEAM
	VALUE_PENALTY_T2
	VALUE_PENALTY_T1_1	//PENALTIES - SKATER
	VALUE_PENALTY_T1_2
	VALUE_PENALTY_T1_3
	VALUE_PENALTY_T1_4
	VALUE_PENALTY_T1_5
	VALUE_PENALTY_T1_6
	VALUE_PENALTY_T1_7
	VALUE_PENALTY_T2_1
	VALUE_PENALTY_T2_2
	VALUE_PENALTY_T2_3
	VALUE_PENALTY_T2_4
	VALUE_PENALTY_T2_5
	VALUE_PENALTY_T2_6
	VALUE_PENALTY_T2_7
	VALUE_PENALTY_T1_J	//PENALTIES - POSITION
	VALUE_PENALTY_T1_B1
	VALUE_PENALTY_T1_B2
	VALUE_PENALTY_T2_J
	VALUE_PENALTY_T2_B1
	VALUE_PENALTY_T2_B2
	VALUE_TIME		//TIME (0 to 60) or special value for timeouts
	VALUE_PERIOD		//PERIOD
	VALUE_JAM		//JAM
	VALUE_T1_J		//LINEUP-SPECIFIC IDS (NUMBER or NUMBER + NAME)
	VALUE_T1_B1
	VALUE_T1_B2
	VALUE_T2_J
	VALUE_T2_B1
	VALUE_T2_B2
	)


//should generate lookup tables for numeric value -> uvs_atlas settings

//internal version, with de-composed tex, texlayer, bits
type SubTexture_ struct{
	uv [4]float32
	tex uint8
	texlayer uint8
	options uint16
}

//external version with encoded selector
type SubTexture struct{
	uv [4]float32
	selector uint32
}


func number_to_uv(value int16) *SubTexture {
	//some kind of lookup that maps numbers to uv texture lookups.
	// for negatives, we also set the selector color select options appropriately to colour it
	var tex = uint8(0) //NUMBERSTEX
	var uv, texlayer, options = [4]float32{0,0,1,1},uint8(0),uint16(0)//lookup into map
	if value < 0 {
		options = options ^ 256 //(option for selecting highlight - checkme)
	}
	return &SubTexture{
			uv: uv,
			selector: uint32(texlayer) + (uint32(tex) << 2) + (uint32(options) << 5),
		}
}

func p_number_to_uv(value uint8, role uint8) *SubTexture {
	//lookup which maps value (1 - 7) and role (NO_ROLE, BLOCKER_ROLE, JAMMER_ROLE) to UV
	var tex = uint8(0) //NUMBERSTEX
	var texlayer = role & 3 //because we arrange the texture atlas so that for "row 0", the lower layers are modifications of the numbers
	var uv = [4]float32{0,0,1,1} //lookup into map
	var options = uint16(0) //
	return &SubTexture {
			uv: uv,
			selector: uint32(texlayer) + (uint32(tex) << 2) + (uint32(options) << 5),
	}
}
