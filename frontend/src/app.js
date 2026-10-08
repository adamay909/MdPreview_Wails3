import { Events, Window } from "@wailsio/runtime";
import { OpenExternal, Html } from "../bindings/mdpreview/mdpreviewer.js";

const bodyElem = document.body;
const contentElem = document.getElementById("content");
const pathElem = document.getElementById("path");
const wcElem = document.getElementById("wordcount");

const name = await Window.Name();

// ---- frontend <-> backend communication
Events.On("updatefile", (e) => updateContent(e.data));

Events.On("newmetadata", (e) => updateMetadata(e.data));

Events.On("themechange", (e) => updateTheme(e.data));

function updateContent(data) {
  if (data.Target != name) {
    return;
  }
  contentElem.innerHTML = data.Html;
  Events.Emit("newplaintext", contentElem.innerText);
}

function updateMetadata(m) {
  if (m.Target != name) {
    return;
  }
  pathElem.textContent = m.ShortPath;
  wcElem.textContent = m.WordCount;
  pathElem.setAttribute("title", m.Path);
}

function updateTheme(t) {
  if (t.Target !== "") {
    if (t.Target !== name) {
      return;
    }
  }
  if (t.Color) {
    if (t.Color === "system") {
      bodyElem.style.removeProperty("color-scheme");
    } else {
      bodyElem.style.setProperty("color-scheme", t.Color);
    }
  }
  if (t.Font) {
    if (t.Font === "serif") {
      contentElem.style.setProperty(
        "font-family",
        '"Merriweather", "Noto Serif", "Yu Mincho", "Hiragino Mincho ProN", "IPAexMincho","Noto Serif JP", serif',
      );
    } else {
      contentElem.style.setProperty(
        "font-family",
        '"Roboto", "Yu Gothic Medium", "Yu Gothic", YuGothic, "Hiragino Kaku Gothic ProN", "Hiragino Sans", Meiryo, sans-serif',
      );
    }
  }
}
// ---- links: prevent navigating to external page
contentElem.addEventListener("click", (e) => {
  const a = e.target.closest("a");
  if (!a) return;
  const href = a.getAttribute("href") || "";
  if (href.startsWith("#")) return;
  e.preventDefault();
  if (href.startsWith("https:") | href.startsWith("mailto:")) {
    OpenExternal(href).catch(() => {});
  }
});

// ---- execute on initial opening of Window
contentElem.innerHTML = await Html();
Events.Emit("newplaintext", contentElem.innerText);
