package main

//action bindings and actions implementations

//this is a very rapidly growing set of actions, it would be simplified by decomposing into 
//  main, and sub types [or, in general, having actions dispatch the glfw event value, and the value of the element activated, to the function]
//			[so, for example, allowing "ACTION_SCORE1_INC" to do differently if the mouse wheel scrolled or we have left or right click; or allowing ACTION_PENALTY_TYPE to all be rolled into one call + the value of the calling element to distinguish]
const (
	ACTION_NULL uint8 = iota
	ACTION_SCORE1_INC
	ACTION_SCORE2_INC
	ACTION_PAUSE_TIME_TOG
	ACTION_TIMEOUT_T1
	ACTION_TIMEOUT_T2
	ACTION_OFFICIAL_TO_TOG
	ACTION_PENALTY_T1
	ACTION_PENALTY_T2
	ACTION_PENALTY_T1_1
	ACTION_PENALTY_T1_2
	ACTION_PENALTY_T1_3
	ACTION_PENALTY_T1_4
	ACTION_PENALTY_T1_5
	ACTION_PENALTY_T1_6
	ACTION_PENALTY_T1_7
	ACTION_PENALTY_T2_1
	ACTION_PENALTY_T2_2
	ACTION_PENALTY_T2_3
	ACTION_PENALTY_T2_4
	ACTION_PENALTY_T2_5
	ACTION_PENALTY_T2_6
	ACTION_PENALTY_T2_7
	ACTION_PENALTY_T1_J
	ACTION_PENALTY_T1_B1
	ACTION_PENALTY_T1_B2
	ACTION_PENALTY_T2_J
	ACTION_PENALTY_T2_B1
	ACTION_PENALTY_T2_B2
	ACTION_PENALTY_TYPE_CONTACT
        ACTION_PENALTY_TYPE_CONTACT_OOP
        ACTION_PENALTY_TYPE_DIRECTION
        ACTION_PENALTY_TYPE_CUTTING
        ACTION_PENALTY_TYPE_SKATING_OOB
        ACTION_PENALTY_TYPE_MULITPLAYER
        ACTION_PENALTY_TYPE_OOP
        ACTION_PENALTY_TYPE_FALSESTART
        ACTION_PENALTY_TYPE_DELAYOFGAME
        ACTION_PENALTY_TYPE_EQUIPMENT
        ACTION_PENALTY_TYPE_JERKFACE
        ACTION_PENALTY_TYPE_MISCONDUCT
	ACTION_PREMATURE_END
	ACTION_WINDOW_VIEW_TOG //start of admin actions 
	ACTION_WINDOW_PANE1_TOG
	ACTION_WINDOW_PANE2_TOG
	ACTION_WINDOW_PANE3_TOG
	ACTION_WINDOW_PANE4_TOG
	ACTION_SCOREBOARD_DISPLAY_TOG
	ACTION_REMOTE_TYPE_TOG
	ACTION_SIMPLE_PENALTY_TOG
)
