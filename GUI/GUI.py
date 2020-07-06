import tkinter as tk

from base64 import b64decode
from json import dumps, load, loads
from sys import exit
from time import sleep, time
from urllib.error import URLError
from urllib.request import urlopen


def check_closing(f):
    def wrapper(self, *args, **kwargs):
        """
        Functions will only attempt to run if the program
        hasn't started to close hopefully resulting in a clean exit
        with no errors in console
        """
        if not self.closing:
            return f(self, *args, **kwargs)
    return wrapper


class GUI(tk.Tk):
    
    WIDTH = 10
    HEIGHT = 10

    REJECTED_MOVES = {
        "r": 37,
        "d": 38,
        "l": 39,
        "u": 40,
    }
    # Stop the snake from going backwards on itself

    def __init__(self, conf, delay=False):
        super().__init__()
        # Init the tkinter inherited class

        self.token = ""

        self.title("Snake")

        self.direction = "r"

        self.conf = conf

        self.url = conf.get("server", "http://localhost:8081")

        self.image = None

        self.lock = False

        self.closing = False

        self.tk_str_var_length = tk.StringVar()

        self.tk_str_var_length.set("Length: 0")

        self.tk_str_var_eaten = tk.StringVar()

        self.tk_str_var_eaten.set("Eaten: 0")

        self.label = tk.Label(
            self,
            image=self.init_grid_img(),
            anchor="center"
        )

        self.length_label = tk.Label(
            self,
            textvariable=self.tk_str_var_length,
            anchor="center"
        )

        self.eaten_label = tk.Label(
            self,
            textvariable=self.tk_str_var_eaten,
            anchor="center"
        )

        self.length_label.pack()

        self.eaten_label.pack()

        self.restart_button = tk.Button(
            self,
            command=self.on_restart_press,
            text="Restart"
        )

        self.restart_button.pack()

        self.bind_all("<Key>", self.key_handler)
        # bind_all over bind as it fixes issues on having to 
        # focus on a specific widget
        self.label.pack()

        self.protocol("WM_DELETE_WINDOW", self.close_window)
        # When user closes window

        self.create_game()

        if delay:
            sleep(3)

        self.clock(self.clock_event)

        # Start the game clock

    @check_closing
    def key_handler(self, evt):
        evt.widget.focus_set()

        if evt.keycode not in self.REJECTED_MOVES.values():
            return

        # If not an arrow key, ignore

        if self.REJECTED_MOVES.get(self.direction, 0) == evt.keycode:
            return
        
        # If the move would cause the snake to go directly back on
        # itself, ignore

        if evt.keycode == 37:
            self.direction = "l"
        elif evt.keycode == 38:
            self.direction = "u"
        elif evt.keycode == 39:
            self.direction = "r"
        elif evt.keycode == 40:
            self.direction = "d"

    @check_closing
    def create_game(self):
        # Generate token and create game on server
        try:
            res = loads(
                urlopen(
                    self.url + "/start",
                    data=dumps(
                        {
                            "width": self.WIDTH,
                            "height": self.HEIGHT
                        }
                    ).encode()
                    # Convert to bytes
                ).read().decode()
                # Convert back to string
            )
            # Parse from JSON structure

            if "error" in res.keys():
                exit(res["error"])

            self.token = res["token"]

        except URLError:
            exit("Unable to connect to server")

    @check_closing
    def clock_event(self):
        try:
            res = loads(
                urlopen(
                    self.url + "/move",
                    data=dumps(
                        {
                            "token": self.token,
                            "direction": self.direction
                        }
                    ).encode()
                ).read().decode()
            )

            if "error" in res.keys():
                print("[WARN] Error", res["error"])
                return

            if not res["alive"]:
                return self.emit_dead(res)
            
            if res["won"]:
                return self.emit_won(res)

            self.update_board(res)

        except URLError:
            exit("Unable to connect to server")

    @check_closing
    def clock(self, func):
        start_time = time() * 1000
        interval_time = 0.5 * 1000

        # Work in milliseconds to allow use of modulus in sub-second
        # scenarios

        while True:
            if self.closing:
                return

            self.update()
            # Update tkinter class to keep listening for events
            if self.lock:
                break
            # Ignore further inputs
                
            query_time = start_time + (
                    interval_time - ((time() * 1000) - start_time)
                    % interval_time
            )
            # Keep line length <= 72
            
            if time() * 1000 >= query_time:
                # Tkinter doesn't tend to play nice with threads
                start_time = time() * 1000
                func()
            # Runs once every half a second

        self.mainloop()
        # Keep window responsive after game over

    @check_closing
    def init_grid_img(self):
        # Creates an image using the server /image method
        default_board = {
            "board_positions": [
                [
                    " " for _ in range(self.WIDTH)
                ] for _ in range(self.HEIGHT)
            ],
            "head_colour": self.conf.get("HeadColour", [
                0,
                0,
                255,
            ]),
            "body_colour": self.conf.get("BodyColour", [
                50,
                130,
                170,
            ]),
            "apple_colour": self.conf.get("AppleColour", [
                255,
                0,
                0,
            ]),
            "background_colour": self.conf.get("BackgroundColour", [
                60,
                170,
                50
            ]),
            "block_height": self.conf.get("BlockHeight", 50),
            "block_width": self.conf.get("BlockWidth", 50),
            "border_colour": self.conf.get("BorderColour", [
                255,
                255,
                255
            ])
        }
        
        try:
            res = loads(
                urlopen(
                    self.url + "/image",
                    data=dumps(default_board).encode(),
                ).read().decode()
            )
            # Makes request, returned in JSON format and parsed

            if "error" in res.keys():
                exit(res.get("error"))

            self.image = tk.PhotoImage(data=b64decode(res["image"]))

            return self.image

        except URLError:
            exit("Unable to connect to server")

    @check_closing
    def update_board(self, board):
        self.tk_str_var_eaten.set("Eaten: " + str(board["eaten"]))
        self.tk_str_var_length.set("Length: " + str(board["length"]))

        try:
            # Attempt to create an image for the updated board
            board = {
                "board_positions": board["board"],
                "head_colour": self.conf.get("HeadColour", [
                    0,
                    0,
                    255,
                ]),
                "body_colour": self.conf.get("BodyColour", [
                    50,
                    130,
                    170,
                ]),
                "apple_colour": self.conf.get("AppleColour", [
                    255,
                    0,
                    0,
                ]),
                "background_colour": self.conf.get("BackgroundColour", [
                    60,
                    170,
                    50
                ]),
                "block_height": self.conf.get("BlockHeight", 50),
                "block_width": self.conf.get("BlockWidth", 50),
                "border_colour": self.conf.get("BorderColour", [
                    255,
                    255,
                    255
                ])
            }

            # Make request
            res = loads(
                urlopen(
                    self.url + "/image",
                    data=dumps(board).encode()
                ).read().decode()
            )

            if "error" in res.keys():
                # Error creating image, use previous
                print("[WARN]", res.get("error"))
                return

        except URLError:
            print("[WARN] Unable to connect to server to create image")
            return

        self.image = tk.PhotoImage(data=b64decode(res["image"]))
        self.label.configure(image=self.image)
        
    @check_closing
    def emit_won(self, res):
        self.lock = True
        # Stop any further requests to server
        w = tk.Label(
            self,
            text="You Won! Final Length " + str(res["length"]),
            anchor="n",
            height=20
        )
        w.pack()

    @check_closing
    def emit_dead(self, res):
        self.lock = True
        # Stop any further requests to server
        d = tk.Label(
            self,
            text="You Died! Final Length " + str(res["length"]),
            anchor="n",
            height=20
        )
        d.pack()

    @check_closing
    def on_restart_press(self):
        self.destroy()

        cls, conf = self.__class__, self.conf

        del self

        cls(conf=conf, delay=False)

    @check_closing
    def close_window(self):
        self.closing = True
        self.destroy()
        del self


if __name__ == "__main__":
    with open("config.json") as c:
        con = load(c)
        # Load from config.json into dict
    app = GUI(con, delay=True)
