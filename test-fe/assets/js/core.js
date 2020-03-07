"use strict";

// these need to be accessed inside more than one function so we'll declare them in the global scope
let container;
let camera;
let renderer;
let scene;
let mesh;
let controls;
let object;
let raycaster;
import * as THREE from "../vendor/threeJs/three.js";
import "../vendor/threeJs/orbital-controls.js";
import "./selection-box.js";
import "./selection-helper.js"

// let data = {
//   "groups": {
//     "group0": {
//       "x": 0.4159240298228717,
//       "y": 0.30985597389307695,
//       "z": 0.11883011642484208
//     },
//     "group4": {
//       "x": 0.05094092855738649,
//       "y": 0.09938172107540882,
//       "z": 0.15130238962656856
//     },
//     "group6": {
//       "x": 0.04992061657675103,
//       "y": 0.11220079703737104,
//       "z": 0.47678027988674454
//     },
//     "group8": {
//       "x": 0.6572444401142241,
//       "y": 0.5328621726559709,
//       "z": 0.7593524595341643
//     }
//   },
//   "dataPoints": {
//     "us.newyork.dc1.be.25": {
//       "groupTag": "group0",
//       "x": 0.7335282404198084,
//       "y": 0.8636391735537683,
//       "z": 0.9548256496989751
//     },
//     "ger.munich.xx2.xe.28": {
//       "groupTag": "group0",
//       "x": 0.3382333151396125,
//       "y": 0.8455760743771621,
//       "z": 0.29128736215957035
//     },
//     "ger.munich.xx2.xe.38": {
//       "groupTag": "group4",
//       "x": 0.5690345363304049,
//       "y": 0.8976976535297126,
//       "z": 0.7191543108783685
//     },
//     "us.newyork.dc1.be.0": {
//       "groupTag": "group6",
//       "x": 0.6327394101398829,
//       "y": 0.9446228737318466,
//       "z": 0.775828199749101
//     },
//     "us.newyork.dc1.be.42": {
//       "groupTag": "group8",
//       "x": 0.47957390033787695,
//       "y": 0.6325982257356839,
//       "z": 0.4774096257260159
//     },
//     "us.newyork.dc1.be.45": {
//       "groupTag": "",
//       "x": 0.57957390033787695,
//       "y": 0.6325982257356839,
//       "z": 0.4774096257260159
//     },
//     "us.newyork.dc1.be.400": {
//       "groupTag": "",
//       "x": 0.80957390033787695,
//       "y": 0.6325982257356839,
//       "z": 0.4774096257260159
//     }
//   }
// }
let socket = new WebSocket("ws:167.172.180.181:6671");
var nodes = {};
var config = {
  X: { normalize: true, index: 3, label: "MEMORY %" },
  Y: { normalize: true, index: 1, label: "NET IN+OUT MB/s" },
  Z: { normalize: true, index: 2, label: "CPU %" },
  Size: { normalize: true, index: 4 },
  Luminocity: { normalize: true, index: 5 },
  Blink: {}
  // UpdateRate: 1000,
  // WantsUpdates: true
};
// call the init function to set everything up
init();

function init() {
  // Get a reference to the container element that will hold our scene
  container = document.querySelector("#scene-container");

  // create a Scene
  scene = new THREE.Scene();

  scene.background = new THREE.Color(0x000b36);

  createCamera();
  createControls();
  createLights();
  createGrid();
  createLines();
  // createShapes();
  createRenderer();
  // createServerList();
  // setInterval(function () {
  // }, 500);
  createLabel(config.Y.label, 0.03, 0.001, 0, 1.04, 0.1, 0, 0.7, 0)
  createLabel(config.X.label, 0.04, 0.001, 1.02, 0, 0, 0, 0, 0)
  createLabel(config.Z.label, 0.04, 0.001, 0, 0, 1.12, 0, 1.6, 0)

  let geometry = new THREE.BoxBufferGeometry(1, 1, 1);
  socket.onopen = function (e) {
    console.log("[open] Connection established, send -> server");
    socket.send(JSON.stringify(config));
  };

  socket.onmessage = function (event) {
    // console.log(`[message] Data received: ${event.data} <- server`);
    // var dpc = JSON.parse(event.data)
    // console.dir(event.data)

    let dataCollections = event.data.split(",");
    dataCollections.forEach(element => {
      let columns = element.split("/");
      let netIn = Number(columns[4]) / 1000000
      let netOut = Number(columns[5]) / 1000000
      nodes[columns[0]] = {
        tag: columns[0],
        1: columns[1], //cpu
        2: columns[2], // disk
        3: columns[3], // memory
        4: netIn, // net in
        5: netOut // net out
      };
      processDataPoint(geometry, nodes[columns[0]])

      // console.log("Node:", columns[0], "cpu: ", columns[1], " disk: ", columns[2], " memory: ", columns[3],  " networkIN: ", columns[4], " networkOUT: ", columns[5])
    });
    // console.log("render")
    // render();
    // createServerList()
  };

  renderer.setAnimationLoop(() => {
    // console.log("render")
    scene.dispose()
    render();
  });
}

var selectionBox = new THREE.SelectionBox(camera, scene);
var helper = new THREE.SelectionHelper(selectionBox, renderer, "selectBox");

function createCamera() {
  camera = new THREE.PerspectiveCamera(
    40, // FOV
    container.clientWidth / container.clientHeight, // aspect ratio
    0.01, // near clipping plane
    100 // far clipping plane
  );

  camera.position.set(1.5, 1, 3)
}

function createControls() {
  controls = new THREE.OrbitControls(camera, container);
}

function createLights() {
  // Create an ambient light in order illuminate the obj from all directions
  const ambientLight = new THREE.HemisphereLight(
    0xddeeff, // sky color
    0x202020, // ground color
    20 // intensity
  );
  // Create a directional light
  const mainLight = new THREE.DirectionalLight(0xffffff, 3.0);

  // Move the light back and up a bit
  mainLight.position.set(10, 10, 10);

  // Remember to add the light to the scene
  scene.add(ambientLight, mainLight);
}

function createGrid() {
  let startPoint = 0;
  let wallMaterial = new THREE.LineBasicMaterial({
    color: 0x8d6722,
    opacity: 0.5,
    transparent: true
  });
  let start = 0.2;
  var i;
  for (i = 0; i < 5; i++) {
    let g1 = new THREE.BufferGeometry();
    g1.addAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array([start, 0, 0, start, 1, 0]), 3)
    );

    let xGridLine = new THREE.Line(g1, wallMaterial);

    scene.add(xGridLine);
    start = start + 0.2;
  }

  start = 0.2;
  var i;
  for (i = 0; i < 5; i++) {
    let g1 = new THREE.BufferGeometry();
    g1.addAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array([0, 0, start, 0, 1, start]), 3)
    );

    let xGridLine = new THREE.Line(g1, wallMaterial);

    scene.add(xGridLine);
    start = start + 0.2;
  }

  start = 0.1;
  var i;
  for (i = 0; i < 10; i++) {
    let g1 = new THREE.BufferGeometry();
    g1.addAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array([0, 0, start, 1, 0, start]), 3)
    );

    let xGridLine = new THREE.Line(g1, wallMaterial);

    scene.add(xGridLine);
    start = start + 0.1;
  }
  start = 0.1;
  var i;
  for (i = 0; i < 10; i++) {
    let g1 = new THREE.BufferGeometry();
    g1.addAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array([start, 0, 0, start, 0, 1]), 3)
    );

    let xGridLine = new THREE.Line(g1, wallMaterial);

    scene.add(xGridLine);
    start = start + 0.1;
  }

  start = 0.1;
  var i;
  for (i = 0; i < 10; i++) {
    let g1 = new THREE.BufferGeometry();
    g1.addAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array([0, start, 0, 1, start, 0]), 3)
    );

    let xGridLine = new THREE.Line(g1, wallMaterial);

    scene.add(xGridLine);
    start = start + 0.1;
  }

  start = 0.1;
  var i;
  for (i = 0; i < 10; i++) {
    let g1 = new THREE.BufferGeometry();
    g1.addAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array([0, start, 0, 0, start, 1]), 3)
    );

    let xGridLine = new THREE.Line(g1, wallMaterial);

    scene.add(xGridLine);
    start = start + 0.1;
  }
}

function createLines() {
  const x_geometry = new THREE.BufferGeometry();
  const y_geometry = new THREE.BufferGeometry();
  const z_geometry = new THREE.BufferGeometry();

  const x_line_points = new Float32Array([0, 0, 0, 0, 0, 1]);
  const y_line_points = new Float32Array([0, 1, 0, 0, 0, 0]);
  const z_line_points = new Float32Array([1, 0, 0, 0, 0, 0]);

  x_geometry.addAttribute(
    "position",
    new THREE.BufferAttribute(x_line_points, 3)
  );
  y_geometry.addAttribute(
    "position",
    new THREE.BufferAttribute(y_line_points, 3)
  );
  z_geometry.addAttribute(
    "position",
    new THREE.BufferAttribute(z_line_points, 3)
  );

  const x_line_material = new THREE.LineBasicMaterial({ color: 0x003eff });
  const x_line = new THREE.Line(x_geometry, x_line_material);

  const y_line_material = new THREE.LineBasicMaterial({ color: 0x003eff });
  const y_line = new THREE.Line(y_geometry, y_line_material);

  const z_line_material = new THREE.LineBasicMaterial({ color: 0x003eff });
  const z_line = new THREE.Line(z_geometry, z_line_material);

  scene.add(x_line);
  scene.add(y_line);
  scene.add(z_line);
}



function createLabel(labelText, size, offset, x, y, z, rx, ry, rz) {
  var loader = new THREE.FontLoader();

  loader.load('gg.json', function (font) {

    var textGeo = new THREE.TextGeometry(labelText, {

      font: font,

      size: size,
      height: 0.001,
      curveSegments: 0,

      bevelThickness: 0,
      bevelSize: 0,
      bevelEnabled: false

    });

    var textMaterial = new THREE.MeshPhongMaterial({ color: 0xff0000 });

    var mesh = new THREE.Mesh(textGeo, textMaterial);
    mesh.position.set(x, y, z);
    mesh.rotation.set(rx, ry, rz)

    scene.add(mesh);

  }, function (dd) { console.dir(dd) });
}

function generateGroup(geometry, groups, totals) {
  for (let groupedServers in groups) {
    object = new THREE.Mesh(
      geometry,
      new THREE.MeshStandardMaterial({ color: 0xff0000 })
    );
    let offset = 0;

    object.position.x = 1 * groups[groupedServers].x - 0.5;
    object.position.y = 1 * groups[groupedServers].y - 0.5;
    object.position.z = 1 * groups[groupedServers].z - 0.5;

    object.material.color.setHex(0x00ff00);

    object.scale.x = 0.03;
    object.scale.y = 0.03;
    object.scale.z = 0.03;
    offset = 0.02;

    let label = "" + totals[groupedServers];

    createLabel(
      label,
      0.006,
      offset,
      object.position.x,
      object.position.y,
      object.position.z
    );

    let objTarget = generateShapeTarget(groupedServers);
    object.userData.group = objTarget;
    object.userData.isSelectable = true;
    object.userData.selected = false;
    object.name = groupedServers;

    // create label for the object
    scene.add(objTarget);
    scene.add(object);
  }
}
function checkDataPointForAlerts(dp, targeting) {
  let shouldUpdateLines = false
  // console.log(dp.position.x, dp.position.y, dp.position.z)
  if (dp.position.y >= 0.9) {
    // console.log("Y:", dp.position.y, dp.name)
    targeting.visible = true;
    shouldUpdateLines = true
    dp.material.emissive = new THREE.Color(0x000000);
  } else {
    // dp.material.emissive = new THREE.Color(0x0000ff);
  }
  if (dp.position.x >= 0.9 || dp.position.x < 0.3) {
    // console.log("X:", dp.position.x, dp.name)
    targeting.visible = true;
    shouldUpdateLines = true

    dp.material.emissive = new THREE.Color(0x000000);
  } else {
    // dp.material.emissive = new THREE.Color(0x0000ff);
  }
  if (dp.position.z >= 0.9) {
    // console.log("Z:", dp.position.z, dp.name)
    targeting.visible = true;
    shouldUpdateLines = true

    dp.material.emissive = new THREE.Color(0x000000);
  } else {
    // dp.material.emissive = new THREE.Color(0x0000ff);
  }

  if (shouldUpdateLines) {
    for (let line of targeting.children) {
      line.material.color = new THREE.Color(0xff0000);
    }
  } else {
    for (let line of targeting.children) {
      line.material.color = new THREE.Color(0xff00ff);
    }
  }

}
let serverList = document.getElementById("all-servers-list");
let setNodes = new Set();
function processDataPoint(geometry, dp) {

  let odp = scene.getObjectByName(dp["tag"])
  if (odp !== undefined) {

    let odpt = scene.getObjectByName(dp["tag"] + "-target")

    odp.position.x = dp[3] / 100;
    odp.position.z = dp[1] / 100;
    odp.position.y = (dp[4] + dp[5]) / 100;
    // if (odp.name === "node-lon1-01") {
    // console.dir(odpt)
    for (let line of odpt.children) {
      if (line.axis === "x") {
        let TargetPoints = new Float32Array([odp.position.x, odp.position.y, odp.position.z, 0, odp.position.y, odp.position.z]);
        line.geometry.attributes.position = new THREE.BufferAttribute(TargetPoints, 3)
      } else if (line.axis === "z") {
        let TargetPoints = new Float32Array([odp.position.x, odp.position.y, odp.position.z, odp.position.x, odp.position.y, 0]);
        line.geometry.attributes.position = new THREE.BufferAttribute(TargetPoints, 3)
      } else if (line.axis === "y") {
        let TargetPoints = new Float32Array([odp.position.x, odp.position.y, odp.position.z, odp.position.x, 0, odp.position.z]);
        line.geometry.attributes.position = new THREE.BufferAttribute(TargetPoints, 3)
      }

    }
    checkDataPointForAlerts(odp, odpt)
    odp.matrixAutoUpdate = false;

    odp.updateMatrix();

    return
  }
  object = new THREE.Mesh(
    geometry,
    new THREE.MeshStandardMaterial({ color: 0xff0000 })
  );
  let targetMaterial = new THREE.LineBasicMaterial({
    color: 0xff00ff,
    // opacity: 0.55,
    // transparent: true
  });
  let targetGroup = new THREE.Group();

  object.position.x = dp[1] / 100;
  object.position.y = dp[2] / 100;
  object.position.z = dp[3] / 100;

  object.material.color.setHex(0x00f0f);

  object.scale.x = 0.01;
  object.scale.y = 0.01;
  object.scale.z = 0.01;

  let xTargetGeometry = new THREE.BufferGeometry();
  let yTargetGeometry = new THREE.BufferGeometry();
  let zTargetGeometry = new THREE.BufferGeometry();

  let xTargetPoints = new Float32Array([
    object.position.x,
    object.position.y,
    object.position.z,
    0,
    object.position.y,
    object.position.z
  ]);

  let yTargetPoints = new Float32Array([
    object.position.x,
    object.position.y,
    object.position.z,
    object.position.x,
    0,
    object.position.z
  ]);

  let zTargetPoints = new Float32Array([
    object.position.x,
    object.position.y,
    object.position.z,
    object.position.x,
    object.position.y,
    0
  ]);

  xTargetGeometry.addAttribute(
    "position",
    new THREE.BufferAttribute(xTargetPoints, 3)
  );
  yTargetGeometry.addAttribute(
    "position",
    new THREE.BufferAttribute(yTargetPoints, 3)
  );
  zTargetGeometry.addAttribute(
    "position",
    new THREE.BufferAttribute(zTargetPoints, 3)
  );

  let xTarget = new THREE.Line(xTargetGeometry, targetMaterial);
  let yTarget = new THREE.Line(yTargetGeometry, targetMaterial);
  let zTarget = new THREE.Line(zTargetGeometry, targetMaterial);
  xTarget.axis = "x"
  zTarget.axis = "z"
  yTarget.axis = "y"
  targetGroup.add(xTarget, yTarget, zTarget);
  targetGroup.name = dp["tag"] + "-target";
  targetGroup.visible = false;

  object.userData.group = targetGroup;
  object.userData.isSelectable = true;
  object.userData.selected = false;
  object.name = dp["tag"];



  // create label for the object
  scene.add(targetGroup);
  scene.add(object);

  let singlePointElement = document.createElement("li");
  let singlePointContainer = document.createElement("span");
  let singlePointElementText = document.createTextNode(object.name);
  singlePointContainer.appendChild(singlePointElementText);
  singlePointElement.appendChild(singlePointContainer);
  serverList.appendChild(singlePointElement);


}

function generateShapeTarget(targetObject) {
  let targetMaterial = new THREE.LineBasicMaterial({
    color: 0xff00ff,
    // opacity: 0.99,
    // transparent: true
  });

  let targetGroup = new THREE.Group();
  let xTargetGeometry = new THREE.BufferGeometry();
  let yTargetGeometry = new THREE.BufferGeometry();
  let zTargetGeometry = new THREE.BufferGeometry();

  let xTargetPoints = new Float32Array([
    object.position.x,
    object.position.y,
    object.position.z,
    -0.5,
    object.position.y,
    object.position.z
  ]);

  let yTargetPoints = new Float32Array([
    object.position.x,
    object.position.y,
    object.position.z,
    object.position.x,
    -0.5,
    object.position.z
  ]);

  let zTargetPoints = new Float32Array([
    object.position.x,
    object.position.y,
    object.position.z,
    object.position.x,
    object.position.y,
    -0.5
  ]);

  xTargetGeometry.addAttribute(
    "position",
    new THREE.BufferAttribute(xTargetPoints, 3)
  );
  yTargetGeometry.addAttribute(
    "position",
    new THREE.BufferAttribute(yTargetPoints, 3)
  );
  zTargetGeometry.addAttribute(
    "position",
    new THREE.BufferAttribute(zTargetPoints, 3)
  );

  let xTarget = new THREE.Line(xTargetGeometry, targetMaterial);
  let yTarget = new THREE.Line(yTargetGeometry, targetMaterial);
  let zTarget = new THREE.Line(zTargetGeometry, targetMaterial);
  targetGroup.add(xTarget, yTarget, zTarget);

  targetGroup.name = targetObject + "-target";
  targetGroup.visible = false;

  return targetGroup;
}



function createRenderer() {
  // Create a WebGLRenderer and set its width and height
  renderer = new THREE.WebGLRenderer({ antialias: true });

  // create a WebGLRenderer and set its width and height
  renderer.setSize(container.clientWidth, container.clientHeight);

  renderer.shadowMap.enabled = true;
  renderer.shadowMap.type = THREE.PCFShadowMap;

  renderer.setPixelRatio(window.devicePixelRatio);

  renderer.physicallyCorrectLights = true;

  renderer.gammaFactor = 2.2;
  renderer.gammaOutput = true;

  // Add the automatically created <canvas> element to the page
  container.appendChild(renderer.domElement);
}

function render() {
  // render, or 'create a still image', of the scene
  // this will create one still image / frame each time the animate
  // function calls itself
  renderer.render(scene, camera);
}

// a function that will be called every time the window gets resized.
// It can get called a lot, so don't put any heavy computation in here!
function onWindowResize() {
  console.log("You resized the browser window!");

  // set the aspect ratio to match the new browser window aspect ratio
  camera.aspect = container.clientWidth / container.clientHeight;

  // update the camera's frustum
  camera.updateProjectionMatrix();

  // update the size of the renderer AND the canvas
  renderer.setSize(container.clientWidth, container.clientHeight);
}

// SELECTION AND TARGETINg

var selectionBox = new THREE.SelectionBox(camera, scene);
var helper = new THREE.SelectionHelper(selectionBox, renderer, "selectBox");
function ctrlPress(event) {
  if (event.ctrlKey) {
    controls.enabled = false;
    helper.enabled = true;
  }
}

function ctrlRelease(event) {
  if (!event.ctrlKey) {
    controls.enabled = true;
    helper.enabled = false;
  }
}

let selectedList = document.getElementById("selected-list");
function onDocumentMouseDown(event) {
  if (!helper.enabled) {
    return;
  }

  for (var item of selectionBox.collection) {
    item.material.emissive = new THREE.Color(0x000000);
    item.userData.group.visible = false;
    item.userData.selected = false;
  }

  while (selectedList.hasChildNodes()) {
    selectedList.removeChild(selectedList.firstChild);
  }

  selectionBox.startPoint.set(
    (event.clientX / window.innerWidth) * 2 - 1,
    -(event.clientY / window.innerHeight) * 2 + 1,
    0.5
  );
}

function onDocumentMouseMove(event) {
  if (!helper.enabled) {
    return;
  }

  if (helper.isDown) {
    for (var i = 0; i < selectionBox.collection.length; i++) {
      selectionBox.collection[i].material.emissive = new THREE.Color(0x000000);
    }

    selectionBox.endPoint.set(
      (event.clientX / window.innerWidth) * 2 - 1,
      -(event.clientY / window.innerHeight) * 2 + 1,
      0.5
    );

    var allSelected = selectionBox.select();
    for (var i = 0; i < allSelected.length; i++) {
      allSelected[i].material.emissive = new THREE.Color(0x0000ff);
      allSelected[i].userData.selected = true;
      allSelected[i].userData.group.visible = true;
    }
  }
}

function onDocumentMouseUp(event) {
  if (!helper.enabled) {
    return;
  }

  selectionBox.endPoint.set(
    (event.clientX / window.innerWidth) * 2 - 1,
    -(event.clientY / window.innerHeight) * 2 + 1,
    0.5
  );

  var allSelected = selectionBox.select();
  for (var i = 0; i < allSelected.length; i++) {
    allSelected[i].material.emissive = new THREE.Color(0x0000ff);
    let node = document.createElement("li");
    let textNode = document.createTextNode(allSelected[i].name);
    node.appendChild(textNode);
    selectedList.appendChild(node);
  }
}

function onSelectedListMouseOver(event) {
  let object = scene.getObjectByName(event.target.innerHTML);
  let objectTarget = scene.getObjectByName(object.userData.group.name);

  object.material.emissive = new THREE.Color(0xffff00);

  for (let line of objectTarget.children) {
    line.material.color = new THREE.Color(0xffff00);
  }
}

function onSelectedListMouseOut(event) {
  let object = scene.getObjectByName(event.target.innerHTML);
  let objectTarget = scene.getObjectByName(object.userData.group.name);

  object.material.emissive = new THREE.Color(0x0000ff);

  for (let line of objectTarget.children) {
    line.material.color = new THREE.Color(0xff00ff);
  }
}

// let serverList = document.getElementById("all-servers-list");
function onServerListMouseOver(event) {
  if (event.target.tagName == "SPAN") {
    let object = scene.getObjectByName(event.target.innerHTML);
    let objectTarget = scene.getObjectByName(object.userData.group.name);

    if (object.userData.selected == false) {
      objectTarget.visible = true;
    }

    object.material.emissive = new THREE.Color(0xffff00);

    for (let line of objectTarget.children) {
      line.material.color = new THREE.Color(0xffff00);
    }
  }
}

function onServerListMouseOut(event) {
  if (event.target.tagName == "SPAN") {
    let object = scene.getObjectByName(event.target.innerHTML);
    let objectTarget = scene.getObjectByName(object.userData.group.name);
    if (object.userData.selected == false) {
      objectTarget.visible = false;
      object.material.emissive = new THREE.Color(0x000000);
    } else {
      object.material.emissive = new THREE.Color(0x0000ff);
    }
    for (let line of objectTarget.children) {
      line.material.color = new THREE.Color(0xff00ff);
    }
  }
}


window.addEventListener("resize", onWindowResize, false);

document.addEventListener("keydown", ctrlPress);
document.addEventListener("keyup", ctrlRelease);

document.addEventListener("mousedown", onDocumentMouseDown);
document.addEventListener("mousemove", onDocumentMouseMove);
document.addEventListener("mouseup", onDocumentMouseUp);

selectedList.addEventListener("mouseover", onSelectedListMouseOver);
selectedList.addEventListener("mouseout", onSelectedListMouseOut);

serverList.addEventListener("mouseover", onServerListMouseOver);
serverList.addEventListener("mouseout", onServerListMouseOut);
