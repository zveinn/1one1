import * as THREE from "./three/three.module.js";
import { OrbitControls } from "./helpers/OrbitControls.js";

//Mock data import
import data from "./data.js";

var camera, scene, renderer, raycaster, controls, mouse, INTERSECTED, axes;
var AXESLENGTH = 14;
var minZoomIn = 10;
var minZoomOut = 80;
var initialZoom = 60;

var CUBEGeometry = new THREE.BoxBufferGeometry(0.2, 0.2, 0.2);
var CUBEMaterial = new THREE.MeshBasicMaterial({
  color: 0xff9900,
  side: THREE.DoubleSide
});

init();
animate();

function init() {
  scene = new THREE.Scene();
  raycaster = new THREE.Raycaster();
  mouse = new THREE.Vector2();

  startRenderer();
  renderCamera();
  createControls();
  addAxesHelper();
  // renderData();
  renderAxeLabel();

  let socket = new WebSocket("ws:127.0.0.1:6671");
  // datapoint = {index:number, value:number}
  var nodes = {};
  var config = {
    X: { normalize: true, index: 1 },
    Y: { normalize: true, index: 2 },
    Z: { normalize: true, index: 3 },
    Size: { normalize: true, index: 4 },
    Luminocity: { normalize: true, index: 5 },
    Blink: {},
    UpdateRate: 1000,
    WantsUpdates: true
  };

  socket.onopen = function(e) {
    console.log("[open] Connection established, send -> server");
    socket.send(JSON.stringify(config));
  };

  socket.onmessage = function(event) {
    console.log(`[message] Data received: ${event.data} <- server`);
    // var dpc = JSON.parse(event.data)
    // console.dir(event.data)
    let dataCollections = event.data.split(",");
    dataCollections.forEach(element => {
      let columns = element.split("/");
      nodes[columns[0]] = {
        tag: columns[0],
        cpu: columns[1],
        disk: columns[2],
        memory: columns[3],
        netin: columns[4],
        netout: columns[5]
      };
      // console.log("Node:", columns[0], "cpu: ", columns[1], " disk: ", columns[2], " memory: ", columns[3],  " networkIN: ", columns[4], " networkOUT: ", columns[5])
    });
    render();
  };

  let cubes = {};

  function render() {
    Object.keys(nodes).forEach(key => {
      if (cubes[nodes[key].tag] === undefined) {
        cubes[nodes[key].tag] = new THREE.Mesh(CUBEGeometry, CUBEMaterial);

        console.log("MAKING A NEW CUBE ...", cubes[nodes[key].tag]);
        scene.add(cubes[nodes[key].tag]);
      }
      cubes[nodes[key].tag].position.set(
        nodes[key].disk / 10,
        nodes[key].cpu / 10,
        nodes[key].memory / 10
      );
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
  document.body.appendChild(renderer.domElement);
}

/**
 * @description Render camera and set it's initial position
 */
function renderCamera() {
  camera = new THREE.PerspectiveCamera(
    initialZoom,
    window.innerWidth / window.innerHeight,
    1,
    1000
  );
  camera.position.set(15, 20, 30);
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

  loader.load("./fonts/helvetiker_regular.typeface.json", function(font) {
    var textMaterial = new THREE.MeshBasicMaterial({ color: 0xffffff });
    var textOptions = {
      font: font,
      size: 0.2,
      height: 0.02
    };

    labels.map(label => {
      var object = new THREE.TextGeometry(
        `${label.ax} ${label.name}`,
        textOptions
      );
      var text = new THREE.Mesh(object, textMaterial);
      text.name = "axes-label";
      text.position[label.ax] = AXESLENGTH;
      scene.add(text);
    });
  });
}

/**
 * @description Render points
 */
function renderData() {
  data.slice(1, 200).map(point => {
    var cube = new THREE.Mesh(CUBEGeometry, CUBEMaterial);
    cube.position.set(
      point.position.x / 10,
      point.position.y / 10,
      point.position.z / 10
    );
    cube.name = point.name;
    scene.add(cube);
  });
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
