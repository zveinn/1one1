# Lynx front end

This is the front-end part of the glorious Lynx app.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

Mission critical packages:
```
nodejs,
ejs,
express,
```

### Installing

Step one - install nodejs
```
sudo apt-get install nodejs
```

Step two - install project dependencies
- navigate to project and run
```
npm install
```

Step 3 - run development mode
```
node server.js
```

## Data structure input

The data should be a JSON type object.
Should be looking like this: 
```
{
    "groups": {
        "group0": {
            "x": 0.4159240298228717,
            "y": 0.30985597389307695,
            "z": 0.11883011642484208
        },
        "group4": {
            "x": 0.05094092855738649,
            "y": 0.09938172107540882,
            "z": 0.15130238962656856
        },
        "group6": {
            "x": 0.04992061657675103,
            "y": 0.11220079703737104,
            "z": 0.47678027988674454
        },
        "group8": {
            "x": 0.6572444401142241,
            "y": 0.5328621726559709,
            "z": 0.7593524595341643
        }
    },
    "dataPoints": {
        "us.newyork.dc1.be.25": {
            "groupTag": "group0",
            "x": 0.7335282404198084,
            "y": 0.8636391735537683,
            "z": 0.9548256496989751
        },
        "ger.munich.xx2.xe.28": {
            "groupTag": "group0",
            "x": 0.3382333151396125,
            "y": 0.8455760743771621,
            "z": 0.29128736215957035
        },
        "ger.munich.xx2.xe.38": {
            "groupTag": "group4",
            "x": 0.5690345363304049,
            "y": 0.8976976535297126,
            "z": 0.7191543108783685
        },
        "us.newyork.dc1.be.0": {
            "groupTag": "group6",
            "x": 0.6327394101398829,
            "y": 0.9446228737318466,
            "z": 0.775828199749101
        },
        "us.newyork.dc1.be.42": {
            "groupTag": "group8",
            "x": 0.47957390033787695,
            "y": 0.6325982257356839,
            "z": 0.4774096257260159
        },
        "us.newyork.dc1.be.45": {
            "groupTag": "",
            "x": 0.57957390033787695,
            "y": 0.6325982257356839,
            "z": 0.4774096257260159
        },
        "us.newyork.dc1.be.400": {
            "groupTag": "",
            "x": 0.80957390033787695,
            "y": 0.6325982257356839,
            "z": 0.4774096257260159
        }
    }
}
```

## Functions

### FUNCTION: createShapes()
input: void
output: void
Is a controler for rendering gorups and individual data points


### FUNCTION: createLabel()
- input: 
    - textLabel: str - how many obj in the group, 
    - size: int - font size,
    - offset: float - recomend to offset on the z axis so the label apears in front of the shape,
    - positionX: float - represents where the object is positioned in space on x axis,
    - positionY: float - represents where the object is positioned in space on y axis,
    - positionZ: float - represents where the object is positioned in space on z axis,
- output: void

Gets where the object is positioned and prints to the screen a label in front of the object containing how many things are in a group


### FUNCTION: generateGroup()
- input: 
    - geometry: obj - that tells the group what shape it will take,
    - gorups: obj - all the objects inside the "groups" key from main data (how many groups there are),
    - totals: int - how many objects are inside a gorup
- output: void

Gets a geometry obj and renders all the group points


### FUNCTION: generateSinglePoint()
- input:
    - geometry: obj - that tells the group what shape it will take,
    - dataPoints: obj - all objects inside the "dataPoints" key from main data(how many points)
- output: void

Gets geometry obj and filters all the points that have an empty groupTag (groupTag === ""), and renders all the points


### FUNCTION: generateShapeTarget()
- input: 
    - targetObject: object
- ouput: target group attached to a specific shape object

This function gets a shape object and attaches vertices to it


### FUNCTION: updateDotPositions()
- input: void
- output: void

This function is deprecated, should be updated. It used to make the dots - when we used sprites instead of full 3d object - move


### FUNCTION: onWindowResize()
- input: void
- output: void

Calculates the height and width of the currect screen, and if changes are made it adapts the canvas to the new width and height - aka responsive screen


### FUNCTION: ctrlPress()
- input: 
    - event: expects a keyboard input - control key at this moment
- output: void

Waits for a control press key, disables obrit controls for the cube, and makes helper enabled


### FUNCTION: ctrlRelease()
- input: 
    -  event: same as ctrlPress() - reffer to it for more details
- output: void

Waits for a control press key, enables orbit controls for the cube, and makes helper disabled

### FUNCTION: onDocumentMouseDown()
- input:
    - event: expects a mouse click event
- output: void

This function iterates over the selected items and resets their color if there is only one click, no drag, makes targets invisible, and sets selected flag to false. 

It is also in charge of despawning the selected list items if only a click is dectected.

If there is a click and drag then spawns a starting point for the select box


### FUNCTION: onDocumentMouseMove()
- input: 
    - event: expects a mouse drag event
- output: void

This function iterates through the things in the selected box, and makes their material "glow" aka emissive, signaling that they are being selected

It also makes the target vertices visible and sets the selected flag to true


### FUNCTION: onDocumentMouseUp()
-  input:
    - event: expects a mouse click release event
- output: void

This function determines where the end point of the select box, iterates through the selected items and creates the selected list.a


### FUNCTION: onServerListMouseOver()
- input: 
    - event: expects a hover event
- output: void

This function checks that only spans are being hovered, and gets the object by it's name which is contained by the span as innerHTML / text.


### FUNCTION: createServerList()
- input: void
- output: void

This function is responsible for generating the server list, depending of the items in the gorups and single data points keys


## Geometries
Geometry documentation
(https://threejs.org/docs/#api/en/core/Geometry)


## Data loading
The data is found in [ProjectHome]/storage/db.json

Data is being loaded from a file inside the index.ejs file - check the format above in the Data structure input section

This is where the data is loaded:
```
<script>
	const data = <%- data %>;
</script>
```

## DB file items generation
Generator file is in [ProjectHome]/storage/db_generator.py

Terminal comands:
```
python3 db_generator.py
```
to generate files

Now it generates one group, with one item inside, and itermitently makes a data point separately.