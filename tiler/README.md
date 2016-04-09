Tiler: move & resize windows on a grid
======================================

This program shows a cute 12x12 grid and lets you move & resize
current window on it, on current screen. Original window position
(rounded to the grid lines) is marked green; selected size/position is
highlighted.

Key Bindings
------------

 - cursor keys, _awsd_: move cursor around
 - _AWSD_: move cursor to top/left/bottom/right edge
 - _Esc_: cancel the operation
 - _Enter_: if selection is not started, starts selection; if
   selection is active, confirms it
 - _Space_: moves selection origin to current position (even if
   selection is already started)
 - _Tab_: switches cursor and selection origin
 - _Backspace_, _q_: removes selection
 - _e_: selects original window position
 - _h_, _v_: maximizes selection horizontally or vertically
 - _x_, _y_: moves to _prefix_ on horizontal/vertical axis
 - _1_–_9_, _0_, _-_, _=_: sets _prefix_ to a number 1–12 (_0_ is 10,
   _-_ is 11, _=_ is 12). If next command is a cursor key or _awsd_
   movement, it will move by prefix (e.g. _3d_ moves 3 fields to the
   right). If next command is a jump (_x_/_y_), it will jump to
   specified column or row (e.g. _-y_ will move to 11th row).
   
