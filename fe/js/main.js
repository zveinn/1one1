import * as THREE from "./three/three.module.js";
import { OrbitControls } from "./helpers/OrbitControls.js";

//Mock data import
import data from "./data.js";

var camera, scene, renderer, raycaster, controls, mouse, INTERSECTED, axes;
var AXESLENGTH = 1;
var minZoomIn = 1;
var minZoomOut = 80;
var initialZoom = 60;

var CUBEGeometry = new THREE.BoxBufferGeometry(0.01, 0.01, 0.01);
var CUBEMaterial = new THREE.MeshBasicMaterial({
  color: 0xff9900,
  side: THREE.DoubleSide
});

let socket = new WebSocket("ws:167.172.180.181:6671");
// datapoint = {index:number, value:number}
var nodes = {};
var config = {
  X: { normalize: true, index: 3 },
  Y: { normalize: true, index: 1 },
  Z: { normalize: true, index: 2 },
  Size: { normalize: true, index: 4 },
  Luminocity: { normalize: true, index: 5 },
  Blink: {}
  // UpdateRate: 1000,
  // WantsUpdates: true
};

init();
animate();
function getID(id) {
  return document.getElementById(id);
}
document
  .getElementById("update-settings")
  .addEventListener("click", function () {
    config.X = { normalize: true, index: getID("x-axis").innerHTML };
  });
document
  .getElementById("list-button")
  .addEventListener("click", function (event) {
    console.dir(event);
    let target = document.getElementById("list");
    if (target.className === "hide") {
      target.className = "list";
    } else {
      target.className = "hide";
    }
  });

document
  .getElementById("settings-button")
  .addEventListener("click", function (event) {
    console.dir(event);
    let target = document.getElementById("settings");
    if (target.className === "hide") {
      target.className = "settings";
    } else {
      target.className = "hide";
    }
  });

function init() {
  scene = new THREE.Scene();
  raycaster = new THREE.Raycaster();
  mouse = new THREE.Vector2();

  startRenderer();
  renderCamera();
  createControls();
  // addAxesHelper();
  // renderData();
  // renderAxeLabel();
  // createLines();
  createLinesx();
  // createGrid();
  loadConfigDefaults();

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
        4: netIn.toString(), // net in
        5: netOut.toString() // net out
      };
      // console.log("Node:", columns[0], "cpu: ", columns[1], " disk: ", columns[2], " memory: ", columns[3],  " networkIN: ", columns[4], " networkOUT: ", columns[5])
    });
    render();
  };

  let cubes = {};
  function loadConfigDefaults() {
    getID("x-axis").innerHTML = config.X.index;
    getID("y-axis").innerHTML = config.Y.index;
    getID("z-axis").innerHTML = config.Z.index;
  }

  function render() {
    let xx = Object.keys(nodes).sort(function (a, b) { return nodes[b][3] - nodes[a][3] })
    // console.dir(xx)
    document.getElementById("list").innerHTML = ""
    xx.forEach(key => {
      if (cubes[nodes[key].tag] === undefined) {
        cubes[nodes[key].tag] = new THREE.Mesh(CUBEGeometry, CUBEMaterial);

        console.log("MAKING A NEW CUBE ...", cubes[nodes[key].tag]);

        scene.add(cubes[nodes[key].tag]);
      } else {
        // COMPARE DATA AND DONT RENDER IF THERE IS NO CHANGE
      }
      let inAndOut = Number(nodes[key][4]) + Number(nodes[key][5])
      // cubes[nodes[key].tag].position.set(
      //   nodes[key][config.X.index] / 100, // X
      //   nodes[key][config.Y.index] / 100, // Y
      //   inAndOut / 100, // Z
      // );

      cubes[nodes[key].tag].position.set(

        nodes[key][config.X.index] / 100,
        inAndOut / 100,
        nodes[key][config.Y.index] / 100,

      );

      // let nodeListItem = document.getElementById(key);
      // if (nodeListItem === null) {
      let nodeListItem = document.createElement("div");
      nodeListItem.id = key;
      nodeListItem.className = "item";
      document.getElementById("list").appendChild(nodeListItem);
      // }



      let tag =
        nodes[key].tag +
        "  >> mem/cpu [ " +
        nodes[key][3] + "/" + nodes[key][1] +
        " ] >> in[ " +
        Math.round(nodes[key][4] * 100) / 100 +
        " ] out[ " +
        Math.round(nodes[key][5] * 100) / 100
        +
        " ]";

      nodeListItem.innerHTML = tag;
    });

  }

  // EVENT LISTENERS
  document.addEventListener("mousemove", onDocumentMouseMove, false);
  document.addEventListener("mousedown", onDocumentMouseDown, false);

  window.addEventListener("resize", onWindowResize, false);
}

/**
 * @description Add axes helper ( x,y,z )
 */
function addAxesHelper() {
  axes = new THREE.AxesHelper(AXESLENGTH);
  axes.position.set(0, 0, 0);
  scene.add(axes);
}

/**
 * @description Create controls for zooming and other
 */
function createControls() {
  controls = new OrbitControls(camera, renderer.domElement);
  controls.minDistance = minZoomIn;
  controls.maxDistance = minZoomOut;
  controls.maxPolarAngle = Math.PI / 2;
}

/**
 * @description Start rendering
 */
function startRenderer() {
  renderer = new THREE.WebGLRenderer({ antialias: true });
  renderer.setPixelRatio(window.devicePixelRatio);
  renderer.setSize(window.innerWidth, window.innerHeight);
  document.getElementById("root").appendChild(renderer.domElement);
}

/**
 * @description Render camera and set it's initial position
 */
function renderCamera() {
  camera = new THREE.PerspectiveCamera(
    initialZoom,
    window.innerWidth / window.innerHeight,
    0.1,
    100
  );
  camera.position.set(1, 1, 2);
  scene.add(camera);
}

/**
 * @description Render axes text label
 */
function renderAxeLabel() {
  var labels = [
    { ax: "x", name: "MEMORY 100%" },
    { ax: "y", name: "CPU 100%" },
    { ax: "z", name: "DISK 100%" }
  ];
  var loader = new THREE.FontLoader();

  // loader.load("./fonts/helvetiker_regular.typeface.json", function(font) {
  console.log("loaded font ...");
  var textMaterial = new THREE.MeshBasicMaterial({ color: 0xffffff });
  var textOptions = {
    font: new THREE.Font(),
    size: 1,
    height: 0.04
  };

  labels.map(label => {
    var object = new THREE.TextGeometry(
      `${label.ax} ${label.name}`,
      textOptions
    );
    var text = new THREE.Mesh(object, textMaterial);
    text.name = "axes-label";
    text.position[label.ax] = 0;
    console.log("adding text");
    scene.add(text);
  });
  // });
  // console.log("redered text..");
}

/**
 * @description On window resize event
 */
function onWindowResize() {
  var width = window.innerWidth;
  var height = window.innerHeight;
  renderer.setSize(width, height);
  camera.aspect = width / height;
  camera.updateProjectionMatrix();
}

/**
 * @description On mouse move event
 * @param event
 */
function onDocumentMouseMove(event) {
  event.preventDefault();
  mouse.x = (event.clientX / window.innerWidth) * 2 - 1;
  mouse.y = -(event.clientY / window.innerHeight) * 2 + 1;
}

/**
 * @description On mouse down event
 * @param event
 */
function onDocumentMouseDown(event) {
  raycaster.setFromCamera(mouse, camera);
  var intersects = raycaster.intersectObjects(scene.children);
  if (intersects.length > 0) {
    if (
      intersects[0].object.isMesh &&
      intersects[0].object.name === "sm-data"
    ) {
      console.log(intersects[0].object);
    }
  }
}

/**
 * @description Rotate axes text in the same direction as the camera
 */
function updateTextRotation() {
  var childrens = scene.children;
  childrens.map(children => {
    if (children.isMesh && children.name === "axes-label") {
      children.setRotationFromEuler(camera.rotation);
    }
  });
}

/**
 * @description Any update logic
 */
function update() {
  raycaster.setFromCamera(mouse, camera);
  var intersects = raycaster.intersectObjects(scene.children);
  if (intersects.length > 0) {
    if (
      intersects[0].object.isMesh &&
      intersects[0].object.name === "sm-data"
    ) {
      if (INTERSECTED !== intersects[0].object) {
        if (INTERSECTED)
          INTERSECTED.material.color.setHex(INTERSECTED.currentHex);
        INTERSECTED = intersects[0].object;
        INTERSECTED.currentHex = INTERSECTED.material.color.getHex();
        INTERSECTED.material.color.setHex(0xffffff);
      }
    }
  } else {
    if (INTERSECTED) INTERSECTED.material.color.setHex(INTERSECTED.currentHex);
    INTERSECTED = null;
  }
  controls.update();
}

/**
 * @description This function will continuously update the scene
 */
function animate() {
  // console.log("animated..");
  // TODO: introduce a sleep here!
  requestAnimationFrame(animate);
  render();
  update();
  updateTextRotation();
}

/**
 * @description This function will render all the content ( scenes, cameras )
 */
function render() {
  renderer.render(scene, camera);
}
// float32 positions on screen
// xyz > xyz
// 1,2,3 > 4,5,6
function createLinesx() {
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

  start = 0.2;
  var i;
  for (i = 0; i < 5; i++) {
    let g1 = new THREE.BufferGeometry();
    g1.addAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array([0, 0, start, 1, 0, start]), 3)
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
      new THREE.BufferAttribute(new Float32Array([start, 0, 0, start, 0, 1]), 3)
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
      new THREE.BufferAttribute(new Float32Array([0, start, 0, 1, start, 0]), 3)
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
      new THREE.BufferAttribute(new Float32Array([0, start, 0, 0, start, 1]), 3)
    );

    let xGridLine = new THREE.Line(g1, wallMaterial);

    scene.add(xGridLine);
    start = start + 0.2;
  }
}
