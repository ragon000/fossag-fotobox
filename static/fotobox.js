(function(){
  console.log("hi");
ws = new WebSocket("ws://"+wshost+"/ws");


ws.onopen = function() {

   // Web Socket is connected, send data uting send()
   ele = document.getElementById("statusindicator");
   ele.style = 'color: green;';
   ele.innerHTML = 'WS Connected';
};
ws.onclose = function() {
   ele = document.getElementById("statusindicator");
   ele.style = 'color: red;';
   ele.innerHTML = 'WS Disconnected';
   alert("WS Disconnected, reload the page")
};

ws.onmessage = (evt)=>{
  var rec = evt.data;
  console.log(evt.data);
  var json = JSON.parse(rec);
  document.getElementById("qr").src = "data:image/png;base64,"+json.qr;
  document.getElementById("svgtext").innerHTML = json.img;
  document.getElementById("img").src = json.img;
};
})();
