var dspImg = new Image;
dspImg.onload = function () {
  syncSSPs();
};
dspImg.src = "http://localhost:5050/pixelSync.gif";
dspImg.width = 1;
dspImg.height = 1;
dspImg.alt = "";

const dspName = "{{.DSPName}}";
const dspCookieName = dspName + "_cookie_id";
const sspSync = [
{{range $i, $d := .SSPs}}{
    {{range $k, $v := $d}}"{{$k}}": "{{$v}}",
    {{end}}
},{{end}}
];

timeoutID = setTimeout(syncSSPs, 300);

function syncSSPs() {
    if (getCookie(dspCookieName) == "") {
      return;
    }
    sspSync.forEach(ssp => {
      var aud = getCookie(dspCookieName);
      var img = new Image;
      img.src = ssp.url+"?dsp_name="+dspName+"&"+ssp.cookie+"="+aud+"&resync="+ssp.resync;
      img.width = 1;
      img.height = 1;
      img.alt = "";
    });
}

function cookieSync(dsp_url) {
    var x = document.createElement("img");
    x.setAttribute("src", dsp_url);
    x.setAttribute("width", "1");
    x.setAttribute("height", "1");
    x.setAttribute("alt", "");
    document.body.appendChild(x);
}

function getCookie(cname) {
  var name = cname + "=";
  var decodedCookie = decodeURIComponent(document.cookie);
  var ca = decodedCookie.split(';');
  for(var i = 0; i <ca.length; i++) {
    var c = ca[i];
    while (c.charAt(0) == ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length);
    }
  }
  return "";
}
