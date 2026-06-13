from dataclasses import dataclass
from time import time
import cv2

@dataclass
class Frame:
    img: cv2.UMat
    timestamp: int

    def update(self, img: cv2.typing.MatLike):
        self.img = img
        self.timestamp = int(time.time()*1000)
        return self

class LineSegment:

    def __init__(self, endings: tuple[int, int, int, int]):
        (self.xb, self.yb), (self.xe, self.ye) = endings
        self.k = (self.yb - self.ye)/(self.xb - self.xe)
        self.b = self.ye - self.k*self.xe

    def points(self):
        if self.yb == self.ye or self.xb == self.xe:
            return []
        step_y = 1 if self.ye > self.yb else -1
        for y in range(self.yb, self.ye, step_y):
            x = int((y-self.b)/self.k)
            yield (y, x) 
    
    @property
    def square_of_length(self):
        return (self.xe - self.xb)**2 + (self.ye - self.yb)

@dataclass
class FrameWithLines:
    frame: Frame
    lines: list[LineSegment]

