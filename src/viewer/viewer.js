/* MB SSTorytime poc/demo code - to be cleaned up and improved by an expert.. */

var API_SERVER = "http://localhost:8080";

const Im3 = 0;
const Im2 = 1;
const Im1 = 2;
const In0 = 3;
const Il1 = 4;
const Ic2 = 5;
const Ie3 = 6;
const ST_TOP = 7;

const STINDICES = [
  "is a property expressed by",
  "is contained by",
  "comes from",
  "is near/smimilar to",
  "leads to",
  "contains",
  "expresses property",
];

// Disable ENTER submission

window.addEventListener("keydown", function (event) {
  if (event.keyCode == 13) {
    event.preventDefault();
    return false;
  }
});

/***********************************************************/

async function DoPage() {
  let requestURL = API_SERVER + "/Orbit";
  let request = new Request(requestURL);
  let response = await fetch(request);
  let mynote = await response.json();

  DoHeader(mynote);
  DoOrbitPanel(mynote); // Start in orbit

  let nav = document.getElementById("search");

  let counter = document.createElement("span");
  counter.innerHTML = 1;
  counter.id = "counter";
  nav.appendChild(counter);
}

/***********************************************************/

function PrintLink(parent, radius, stindex, arrow, str, nclass, ncptr, ctx) {
  if (arrow == null) {
    arrow = "broken arrow";
  }

  let box = document.createElement("div");
  box.id = "radius-" + radius;

  if (str.includes("\n")) {
    // arrow
    let arrow_link = document.createElement("a");
    arrow_link.textContent = `( ${arrow} )  `;
    //arrow_link.className = 'arrow'+stindex;
    arrow_link.id = "arrow-" + stindex;
    arrow_link.title = STINDICES[stindex];
    arrow_link.class = "tooltip";
    arrow_link.style.fontFamily = "Verdana";
    box.appendChild(arrow_link);

    // pre formatted text
    let text_link = document.createElement("a");
    text_link.onclick = function () {
      sendlinkData(nclass, ncptr);
    };

    let pre = document.createElement("pre");
    pre.textContent = str;
    text_link.appendChild(pre);
    text_link.className = "text";
    box.appendChild(text_link);
  } else {
    // arrow
    let arrow_link = document.createElement("a");
    arrow_link.textContent = `( ${arrow} )  `;

    box.appendChild(arrow_link);
    //arrow_link.className = 'arrow'+stindex;
    arrow_link.id = "arrow-" + stindex;
    arrow_link.title = STINDICES[stindex];
    arrow_link.class = "tooltip";
    arrow_link.style.fontFamily = "Verdana";

    // plain text
    let text_link = document.createElement("a");

    if (IsURL(str, arrow)) {
      text_link.href = str;
      text_link.target = "_blank";
      text_link.rel = "noopener";
    } else if (IsImage(str, arrow)) {
      let img = document.createElement("img");
      img.src = str;
      box.appendChild(img);
    } else {
      text_link.onclick = function () {
        sendlinkData(nclass, ncptr);
      };
    }
    text_link.textContent = str;
    text_link.className = "text";
    text_link.style.fontFamily = "Times";
    if (str.length < 20) {
      text_link.style.fontSize = "200%";
      text_link.style.padding = "10px";
    }
    box.appendChild(text_link);
    if (ctx.length > 0) {
      let cntx = document.createElement("i");
      cntx.id = "cntxt";
      cntx.textContent = " in " + ctx;
      box.appendChild(cntx);
    }
    let br = document.createElement("br");
    box.appendChild(br);
  }

  parent.appendChild(box);

  return parent;
}

/***********************************************************/

function IsImage(str, arrow) {
  if (arrow == "has image" || arrow == "is an image for") {
    if (str.slice(0, 6) == "http:/" || str.slice(0, 7) == "https:/") {
      return true;
    }
  }

  return false;
}

/***********************************************************/

function IsMath(str) {
  if (str.includes("\(") && str.includes("\)")) {
    return true;
  }

  return false;
}

/***********************************************************/

function IsURL(str, arrow) {
  if (arrow == "has URL") {
    if (str.slice(0, 6) == "http:/" || str.slice(0, 7) == "https:/") {
      return true;
    }
  }

  return false;
}

/***********************************************************/

function PrintPath(parent, array) {
  if (array.length < 1) {
    return parent;
  }

  for (let path = 0; path < array.length; path++) {
    if (array[path] == null) {
      continue;
    }

    for (let i = 0; i < array[path].length; i++) {
      if (i % 2 == 0) {
        // node
        let str = array[path][i].Name;
        let ncptr = array[path][i].NCPtr;
        let nclass = array[path][i].NClass;

        if (array[path][i].NPtr != null) {
          ncptr = array[path][i].NPtr.CPtr;
          nclass = array[path][i].NPtr.Class;
        }

        if (str.includes("\n")) {
          let text_link = document.createElement("a");
          text_link.onclick = function () {
            sendlinkData(nclass, ncptr);
          };
          let pre = document.createElement("pre");
          pre.textContent = str;
          text_link.appendChild(pre);
          parent.appendChild(text_link);
        } else {
          let text_link = document.createElement("a");
          text_link.onclick = function () {
            sendlinkData(nclass, ncptr);
          };
          let text = document.createElement("i");
          text.textContent = str;
          text.style.fontFamily = "Times New Roman";

          if (str.length < 20) {
            text.style.fontSize = "150%";
            text.style.padding = "10px";
          }

          text_link.appendChild(text);
          parent.appendChild(text_link);
        }
      } // arrow
      else {
        const then = 2; // reserved vectors
        const prev = 3;

        let arrow = array[path][i].Name;
        let arrptr = array[path][i].Arr;
        let stindex = array[path][i].STindex;

        if (arrptr == then || arrptr == prev) {
          // represent a privileged sequence for proper time
          let hr = document.createElement("p");
          parent.appendChild(hr);
        }

        let arrow_link = document.createElement("a");
        arrow_link.textContent = `( ${arrow} )  `;
        arrow_link.id = `arrow-` + stindex;
        arrow_link.class = "tooltip";
        arrow_link.title = STINDICES[stindex];
        arrow_link.style.fontFamily = "Verdana";

        //arrow_link.className = 'arrow';
        //arrow_link.style.href = 'http...';
        parent.appendChild(arrow_link);
      }
    }

    let spacer = document.createElement("hr");
    parent.appendChild(spacer);
  }

  return parent;
}

/***********************************************************/

function DoHeader(obj) {
  let clearscreen = document.querySelector("article");
  clearscreen.innerHTML = "";

  let header = document.querySelector("header");
  let titlebar = document.createElement("h2");
  titlebar.id = "header_root";

  let title = "app";

  console.log(obj);

  if (obj.events != null) {
    if (obj.events[0] != null) {
      if (obj.events[0].Title != null) {
        title = obj.events[0].Title;
      }
      if (obj.events[0].Text != null) {
        title = obj.events[0].Text;
      }
    }
  } else if (obj.paths != null) {
    if (obj.paths[0] != null) {
      title = obj.paths[0].Title;
    }
  } else {
    title = obj.chapter + " :: " + obj.context + " :: ";
  }

  if (title.length < 60 || IsMath(title)) {
    titlebar.textContent = title;
  } else {
    titlebar.textContent = title.slice(0.6) + "...";
  }

  titlebar.style.fontSize = "70%";

  header.appendChild(titlebar);
}

/***********************************************************/

function DoOrbitPanel(obj) {
  let section = document.querySelector("article");
  let panel = document.createElement("span");
  panel.id = "main_root";
  section.appendChild(panel);

  // List of events unrelated

  if (obj == null) {
    return;
  }

  const separates = 0;

  for (let event of obj.events) {
    ShowEvent(panel, event, separates, "all", "", "h1");
    let hr = document.createElement("hr");
    panel.appendChild(hr);
  }
}

/***********************************************************/

function ShowEvent(panel, event, counter, direction, skiparrow, anchortag) {
  let child = document.createElement("content");

  if (event == null) {
    return;
  }

  let text = counter + ". " + event.Text;

  if (counter == 0) {
    text = "* " + event.Text;
  }

  if (event.Text.includes("\n")) {
    let from_link = document.createElement("a");
    from_link.onclick = function () {
      sendlinkData(event.NPtr.Class, event.NPtr.CPtr);
    };

    let from_text = document.createElement("pre");
    from_text.nameClass = "text";
    from_text.textContent = text; // event.Text;
    from_link.appendChild(from_text);
    from_link.nameClass = "text";
    child.appendChild(from_link);
  } else {
    let from_link = document.createElement("a");
    from_link.onclick = function () {
      sendlinkData(event.NClass, event.NCPtr);
    };
    let from_text = document.createElement(anchortag);
    from_link.appendChild(from_text);
    from_link.nameClass = "text";
    child.appendChild(from_link);

    if (event.Text.length > 90) {
      from_text.style.fontSize = "100%";
    }

    if (event.Text.length > 90 && !IsMath(event.Text)) {
      from_text.textContent = event.Text.slice(0, 70) + "...";
      let small_tot_text = document.createElement("div");
      small_tot_text.textContent = text; // event.Text;
      small_tot_text.id = "full-text";
      child.appendChild(small_tot_text);
    } else {
      from_text.textContent = text; // event.Text;
    }
  }

  let chapter = document.createElement("div");
  chapter.textContent = event.Chap;
  chapter.id = "chapter";
  child.appendChild(chapter);

  // The order is important. Start with coming from

  if (direction != "fwd" && event.Orbits[Im1] != null) {
    for (let ngh of event.Orbits[Im1]) {
      child = PrintLink(
        child,
        ngh.Radius,
        ngh.STindex,
        ngh.Arrow,
        ngh.Text,
        ngh.Dst.Class,
        ngh.Dst.CPtr,
        ngh.Ctx,
      );
      panel.appendChild(child);
    }
  }

  if (event.Orbits[In0] != null) {
    for (let ngh of event.Orbits[In0]) {
      if (skiparrow != ngh.Arrow) {
        child = PrintLink(
          child,
          ngh.Radius,
          ngh.STindex,
          ngh.Arrow,
          ngh.Text,
          ngh.Dst.Class,
          ngh.Dst.CPtr,
          ngh.Ctx,
        );
        panel.appendChild(child);
      }
    }
  }

  if (direction != "fwd" && event.Orbits[Im3] != null) {
    for (let ngh of event.Orbits[Im3]) {
      if (skiparrow != ngh.Arrow) {
        child = PrintLink(
          child,
          ngh.Radius,
          ngh.STindex,
          ngh.Arrow,
          ngh.Text,
          ngh.Dst.Class,
          ngh.Dst.CPtr,
          ngh.Ctx,
        );
        panel.appendChild(child);

        if (IsImage(event.Text, ngh.Arrow)) {
          let img = document.createElement("img");
          img.src = event.Text;
          panel.appendChild(img);
        }
      }
    }
  }

  if (event.Orbits[Ie3] != null) {
    for (let ngh of event.Orbits[Ie3]) {
      if (skiparrow != ngh.Arrow) {
        child = PrintLink(
          child,
          ngh.Radius,
          ngh.STindex,
          ngh.Arrow,
          ngh.Text,
          ngh.Dst.Class,
          ngh.Dst.CPtr,
          ngh.Ctx,
        );
        panel.appendChild(child);
      }
    }
  }

  if (direction != "fwd" && event.Orbits[Im2] != null) {
    for (let ngh of event.Orbits[Im2]) {
      if (skiparrow != ngh.Arrow) {
        child = PrintLink(
          child,
          ngh.Radius,
          ngh.STindex,
          ngh.Arrow,
          ngh.Text,
          ngh.Dst.Class,
          ngh.Dst.CPtr,
          ngh.Ctx,
        );
        panel.appendChild(child);
      }
    }
  }

  if (event.Orbits[Ic2] != null) {
    for (let ngh of event.Orbits[Ic2]) {
      if (skiparrow != ngh.Arrow) {
        child = PrintLink(
          child,
          ngh.Radius,
          ngh.STindex,
          ngh.Arrow,
          ngh.Text,
          ngh.Dst.Class,
          ngh.Dst.CPtr,
          ngh.Ctx,
        );
        panel.appendChild(child);
      }
    }
  }

  if (event.Orbits[Il1] != null) {
    for (let ngh of event.Orbits[Il1]) {
      if (skiparrow != ngh.Arrow) {
        child = PrintLink(
          child,
          ngh.Radius,
          ngh.STindex,
          ngh.Arrow,
          ngh.Text,
          ngh.Dst.Class,
          ngh.Dst.CPtr,
          ngh.Ctx,
        );
        panel.appendChild(child);
      }
    }
  }
}

/***********************************************************/

function DoEntireConePanel(obj) {
  let section = document.querySelector("article");
  let panel = document.createElement("div");
  panel.id = "main_root";
  section.appendChild(panel);

  for (let nptr of obj.paths) {
    let nclass = nptr.NClass;
    let ncptr = nptr.NCPtr;
    let item = document.createElement("p");
    let link = document.createElement("a");
    link.textContent = nptr.Title;
    link.onclick = function () {
      sendlinkData(nclass, ncptr);
    };
    item.appendChild(link);
    let parent = document.createElement("content");
    parent = PrintPath(parent, nptr.Entire);
    panel.appendChild(parent);

    let tab = document.createElement("table");
    let row = document.createElement("tr");
    let col1 = document.createElement("td");
    let hd1 = document.createElement("h2");
    hd1.textContent = "Between Centrality";
    col1.appendChild(hd1);
    let lst1 = document.createElement("ol");

    for (let centrality of nptr.BTWC) {
      let li = document.createElement("li");
      li.textContent = centrality;
      lst1.appendChild(li);
    }
    col1.appendChild(lst1);

    let col2 = document.createElement("td");
    let hd2 = document.createElement("h2");
    hd2.textContent = "Supernodes";
    col2.appendChild(hd2);
    let lst2 = document.createElement("ol");

    for (let snode of nptr.Supernodes) {
      let li = document.createElement("li");
      li.textContent = snode;
      lst2.appendChild(li);
    }
    col2.appendChild(lst2);

    row.appendChild(col1);
    row.appendChild(col2);

    tab.appendChild(row);
    panel.appendChild(tab);

    let hr = document.createElement("hr");
    panel.appendChild(hr);
  }
}

/***********************************************************/

function DoSeqPanel(obj) {
  let section = document.querySelector("article");
  let panel = document.createElement("div");
  panel.id = "main_root";
  section.appendChild(panel);

  // If more than one possible match, then list titles only

  if (obj != null && obj.events != null && obj.events.length > 1) {
    let counter = 1;

    for (let story of obj.events) {
      let link = document.createElement("a");
      let item = document.createElement("h1");
      item.textContent = counter + ". " + story.Text;
      link.onclick = function () {
        sendlinkData(story.ContainNPtr.Class, story.ContainNPtr.CPtr);
      };
      link.appendChild(item);
      panel.appendChild(link);

      let cntx = document.createElement("i");
      cntx.id = "cntxt";
      cntx.textContent = " in " + story.Arrow;
      panel.appendChild(cntx);
      let hr = document.createElement("hr");
      panel.appendChild(hr);
      counter++;
    }
  } else if (obj != null && obj.events != null) {
    // show one full story

    let story = obj.events[0];

    let item = document.createElement("h1");
    item.textContent = story.Text;
    panel.appendChild(item);

    let cntx = document.createElement("i");
    cntx.id = "cntxt";
    cntx.textContent = story.Arrow;
    panel.appendChild(cntx);

    if (story.Axis != null) {
      let counter = 1;
      for (let event of story.Axis) {
        ShowEvent(panel, event, counter, "fwd", "then", "p");
        counter++;
      }
    }
  }
}

/***********************************************************/

function DoBrowsePanel(obj) {
  let section = document.querySelector("article");
  let panel = document.createElement("div");
  panel.id = "main_root";
  section.appendChild(panel);

  for (let nptr of obj.nptrs) {
    let nclass = nptr.NClass;
    let ncptr = nptr.NCPtr;
    let item = document.createElement("p");
    let link = document.createElement("a");
    link.onclick = function () {
      sendlinkData(nclass, ncptr);
    };
    link.textContent = nptr.Title;
    item.appendChild(link);

    item = PrintPath(item, nptr.Il1);
    panel.appendChild(item);
    item = PrintPath(item, nptr.Im1);
    panel.appendChild(item);
    item = PrintPath(item, nptr.Im2);
    panel.appendChild(item);
    item = PrintPath(item, nptr.Ic2);
    panel.appendChild(item);
    item = PrintPath(item, nptr.Ie3);
    panel.appendChild(item);
    item = PrintPath(item, nptr.Im3);
    panel.appendChild(item);
    item = PrintPath(item, nptr.In0);
    panel.appendChild(item);
  }
  let spacer = document.createElement("hr");
  panel.appendChild(spacer);
}

/***********************************************************/

function DoTOCPanel(obj) {
  let section = document.querySelector("article");
  let panel = document.createElement("div");
  panel.id = "main_root";
  section.appendChild(panel);

  for (let chp of obj.TOC) {
    let link = document.createElement("a");
    let item = document.createElement("h1");
    link.onclick = function () {
      sendTOCLinkData(chp.Chapter, "");
    };
    item.textContent = chp.Chapter;
    link.appendChild(item);
    panel.appendChild(link);

    let chap = document.createElement("p");
    chap.id = "toc-panel";

    for (let ctx of chp.Contexts) {
      let link = document.createElement("a");
      let sitem = document.createElement("div");
      sitem.textContent = ctx;
      sitem.id = "toc-elem";
      link.onclick = function () {
        sendTOCLinkData(chp.Chapter, ctx);
      };
      link.appendChild(sitem);
      chap.appendChild(link);
    }

    panel.appendChild(chap);
    let spacer = document.createElement("hr");
    panel.appendChild(spacer);
  }
}

/***********************************************************/
// handlers
/***********************************************************/

function OrbitHandler() {
  let form = document.querySelector("#search");

  async function sendorbitData() {
    let formData = new FormData(form);

    fetch(API_SERVER + "/Orbit", { method: "POST", body: formData })
      .then((response) => {
        if (!response.ok) {
          throw new Error("network returns error");
        }

        return response.json();
      })
      .then((resp) => {
        let prevh = document.getElementById("header_root");

        if (prevh != null) {
          prevh.remove();
        }

        let prevm = document.getElementById("article");

        if (prevm != null) {
          prevm.remove();
        }

        DoHeader(resp);
        DoOrbitPanel(resp);
        MathJax.typeset();
      })

      .catch((error) => {
        // Handle error
        console.log("error ", error);
        let section = document.querySelector("article");
        let text = document.createElement("h2");
        section.textContent = "No results in orbit (perhaps no connection)";
        section.id = "errormesg";
        section.appendChild(text);
      });
  }

  // Take over form submission
  (button = document.getElementById("orbitsubmit")),
    button.addEventListener("click", (event) => {
      event.preventDefault();
      sendorbitData();
    });
}

/***********************************************************/

function ConeHandler() {
  let form = document.querySelector("#search");

  async function sendconeData() {
    let formData = new FormData(form);

    fetch(API_SERVER + "/Cone", { method: "POST", body: formData })
      .then((response) => {
        if (!response.ok) {
          throw new Error("network returns error");
        }

        return response.json();
      })
      .then((resp) => {
        let prevh = document.getElementById("header_root");

        if (prevh != null) {
          prevh.remove();
        }

        let prevm = document.getElementById("article");

        if (prevm != null) {
          prevm.remove();
        }

        DoHeader(resp);
        //DoConePanel(resp);
        DoEntireConePanel(resp);
        MathJax.typeset();
      })

      .catch((error) => {
        // Handle error
        console.log("error ", error);
        let section = document.querySelector("article");
        let text = document.createElement("h2");
        section.textContent = "No results in Geometry (perhaps no connection)";
        section.id = "errormesg";
        section.appendChild(text);
      });
  }

  // Take over form submission
  (button = document.getElementById("pathsubmit")),
    button.addEventListener("click", (event) => {
      event.preventDefault();
      sendconeData();
    });
}

/***********************************************************/

function SeqHandler() {
  let form = document.querySelector("#search");

  async function sendseqData() {
    let formData = new FormData(form);

    fetch(API_SERVER + "/Sequence", { method: "POST", body: formData })
      .then((response) => {
        if (!response.ok) {
          throw new Error("network returns error");
        }

        return response.json();
      })
      .then((resp) => {
        let prevh = document.getElementById("header_root");

        if (prevh != null) {
          prevh.remove();
        }

        let prevm = document.getElementById("article");

        if (prevm != null) {
          prevm.remove();
        }

        DoHeader(resp);
        DoSeqPanel(resp);
        MathJax.typeset();
      })

      .catch((error) => {
        // Handle error
        console.log("error ", error);
        let section = document.querySelector("article");
        let text = document.createElement("h2");
        section.textContent = "No results in Tales (perhaps no connection)";
        section.id = "errormesg";
        section.appendChild(text);
      });
  }

  // Take over form submission
  (button = document.getElementById("seqsubmit")),
    button.addEventListener("click", (event) => {
      event.preventDefault();
      sendseqData();
    });
}

/***********************************************************/

function BrowseHandler() {
  const form = document.querySelector("#search");

  async function sendbrowseData() {
    let formData = new FormData(form);
    document.getElementById("counter").innerHTML = 1;

    fetch(API_SERVER + "/Browse", { method: "POST", body: formData })
      .then((response) => {
        if (!response.ok) {
          throw new Error("network returns error");
        }

        return response.json();
      })
      .then((resp) => {
        DoHeader(resp);
        DoBrowsePanel(resp);
        MathJax.typeset();
      })

      .catch((error) => {
        // Handle error
        console.log("error ", error);

        let section = document.querySelector("article");
        let text = document.createElement("h2");
        section.textContent = "No results in browsing";
        section.id = "errormesg";
        section.appendChild(text);
      });
  }

  // Take over form submission
  (button = document.getElementById("browsesubmit")),
    button.addEventListener("click", (event) => {
      event.preventDefault();
      sendbrowseData();
    });
}

/***********************************************************/

function IncHandler() {
  async function sendincData() {
    let form = document.getElementById("search");
    let formData = new FormData(form);

    // Update *****
    let pagenr = document.getElementById("counter").innerHTML;
    pagenr++;
    document.getElementById("counter").innerHTML = pagenr;

    formData.set("pagenr", pagenr);

    fetch(API_SERVER + "/Browse", { method: "POST", body: formData })
      .then((response) => {
        if (!response.ok) {
          throw new Error("network returns error");
        }

        return response.json();
      })
      .then((resp) => {
        console.log("CHECK INC", JSON.stringify(resp, null, 2));

        let prevh = document.getElementById("header_root");

        if (prevh != null) {
          prevh.remove();
        }

        let prevm = document.getElementById("article");

        if (prevm != null) {
          prevm.remove();
        }

        DoHeader(resp);
        DoBrowsePanel(resp);
      })

      .catch((error) => {
        // Handle error
        console.log("error ", error);
      });
  }

  (button = document.getElementById("inc")),
    button.addEventListener("click", (event) => {
      event.preventDefault();
      sendincData();
    });
}

/***********************************************************/

function DecHandler() {
  async function senddecData() {
    let form = document.getElementById("search");
    let formData = new FormData(form);

    // Update *****
    let pagenr = document.getElementById("counter").innerHTML;

    if (pagenr != 1) {
      pagenr--;
    }

    document.getElementById("counter").innerHTML = pagenr;
    formData.set("pagenr", pagenr);

    fetch(API_SERVER + "/Browse", { method: "POST", body: formData })
      .then((response) => {
        if (!response.ok) {
          throw new Error("network returns error");
        }

        return response.json();
      })
      .then((resp) => {
        console.log("CHECK DEC", JSON.stringify(resp, null, 2));

        let prevh = document.getElementById("header_root");

        if (prevh != null) {
          prevh.remove();
        }

        let prevm = document.getElementById("article");

        if (prevm != null) {
          prevm.remove();
        }

        DoHeader(resp);
        DoBrowsePanel(resp);
      })

      .catch((error) => {
        // Handle error
        console.log("error ", error);
      });
  }

  (button = document.getElementById("dec")),
    button.addEventListener("click", (event) => {
      event.preventDefault();
      senddecData();
    });
}

/***********************************************************/

async function sendlinkData(nclass, ncptr) {
  let formData = new FormData();
  formData.set("nclass", nclass);
  formData.set("ncptr", ncptr);

  fetch(API_SERVER + "/Orbit", { method: "POST", body: formData })
    .then((response) => {
      if (!response.ok) {
        throw new Error("network returns error");
      }

      return response.json();
    })
    .then((resp) => {
      console.log("CHECK ORBIT", JSON.stringify(resp, null, 2));

      let prevh = document.getElementById("header_root");

      if (prevh != null) {
        prevh.remove();
      }

      let prevm = document.getElementById("article");

      if (prevm != null) {
        prevm.remove();
      }

      DoHeader(resp);
      DoOrbitPanel(resp);
      MathJax.typeset();
    })

    .catch((error) => {
      // Handle error
      console.log("error ", error);
    });
}

/***********************************************************/

async function sendTOCLinkData(chap, ctx) {
  let formData = new FormData();
  formData.set("chapter", chap);
  formData.set("contexts", ctx);

  fetch(API_SERVER + "/Browse", { method: "POST", body: formData })
    .then((response) => {
      if (!response.ok) {
        throw new Error("network returns error");
      }

      return response.json();
    })
    .then((resp) => {
      console.log("CHECK TOC", JSON.stringify(resp, null, 2));

      let prevh = document.getElementById("header_root");

      if (prevh != null) {
        prevh.remove();
      }

      let prevm = document.getElementById("article");

      if (prevm != null) {
        prevm.remove();
      }

      DoHeader(resp);
      DoBrowsePanel(resp);
      MathJax.typeset();
    })

    .catch((error) => {
      // Handle error
      console.log("error ", error);
    });
}

/***********************************************************/

function TOCHandler() {
  let form = document.querySelector("#search");

  async function sendTOCData() {
    let formData = new FormData(form);

    fetch(API_SERVER + "/TOC", { method: "POST", body: formData })
      .then((response) => {
        if (!response.ok) {
          throw new Error("network returns error");
        }

        return response.json();
      })
      .then((resp) => {
        let prevh = document.getElementById("header_root");

        if (prevh != null) {
          prevh.remove();
        }

        let prevm = document.getElementById("article");

        if (prevm != null) {
          prevm.remove();
        }

        DoHeader(resp);
        DoTOCPanel(resp);
      })

      .catch((error) => {
        // Handle error
        console.log("error ", error);
        let section = document.querySelector("article");
        let text = document.createElement("h2");
        section.textContent = "No result in TOC";
        section.id = "errormesg";
        section.appendChild(text);
      });
  }

  // Take over form submission
  (button = document.getElementById("toc-submit")),
    button.addEventListener("click", (event) => {
      event.preventDefault();
      sendTOCData();
    });
}

/***********************************************************/

DoPage();
OrbitHandler();
ConeHandler();
SeqHandler();
BrowseHandler();
IncHandler();
DecHandler();
TOCHandler();
