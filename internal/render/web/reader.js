(function () {
  "use strict";

  var sections = Array.prototype.slice.call(document.querySelectorAll(".doc-section"));
  var docButtons = Array.prototype.slice.call(document.querySelectorAll(".doc-item"));
  var tabButtons = Array.prototype.slice.call(document.querySelectorAll(".sidebar-tab"));
  var tabPanels = Array.prototype.slice.call(document.querySelectorAll(".tab-panel"));
  var toc = document.getElementById("toc");
  var titleEl = document.getElementById("doc-title");
  var menuButton = document.getElementById("menu-toggle");
  var themeQuery = window.matchMedia ? window.matchMedia("(prefers-color-scheme: dark)") : null;
  var activeHeadings = [];

  function applyChromaTheme() {
    var dark = !!(themeQuery && themeQuery.matches);
    var blocks = document.querySelectorAll(".chroma");
    for (var i = 0; i < blocks.length; i++) {
      blocks[i].classList.toggle("dark", dark);
      blocks[i].classList.toggle("light", !dark);
    }
  }

  function currentSection() {
    for (var i = 0; i < sections.length; i++) {
      if (!sections[i].hidden) {
        return sections[i];
      }
    }
    return sections[0] || null;
  }

  function collectHeadings(section) {
    activeHeadings = section
      ? Array.prototype.slice.call(section.querySelectorAll("h1, h2, h3"))
      : [];
    updateActiveTOC();
  }

  function updateActiveTOC() {
    if (!toc || activeHeadings.length === 0) {
      return;
    }
    var current = activeHeadings[0].id;
    for (var i = 0; i < activeHeadings.length; i++) {
      var top = activeHeadings[i].getBoundingClientRect().top;
      if (top <= 130) {
        current = activeHeadings[i].id;
      } else {
        break;
      }
    }
    var links = toc.querySelectorAll(".toc-link");
    for (var j = 0; j < links.length; j++) {
      links[j].classList.toggle("active", links[j].getAttribute("data-toc-id") === current);
    }
  }

  function activateTab(name) {
    if (tabButtons.length === 0) {
      return;
    }
    for (var i = 0; i < tabButtons.length; i++) {
      var selected = tabButtons[i].getAttribute("data-tab") === name;
      tabButtons[i].classList.toggle("active", selected);
      tabButtons[i].setAttribute("aria-selected", selected ? "true" : "false");
    }
    for (var j = 0; j < tabPanels.length; j++) {
      tabPanels[j].hidden = tabPanels[j].getAttribute("data-tab-panel") !== name;
    }
    if (name === "toc") {
      updateActiveTOC();
    }
  }

  function replaceTOC(section) {
    if (!toc) {
      return;
    }
    toc.replaceChildren();
    if (!section) {
      return;
    }
    var template = section.querySelector("template.toc-template");
    if (template) {
      toc.appendChild(template.content.cloneNode(true));
    }
  }

  function setHash(docID, anchor) {
    var hash = "#doc=" + encodeURIComponent(docID);
    if (anchor) {
      hash += "&anchor=" + encodeURIComponent(anchor);
    }
    try {
      history.replaceState(null, "", hash);
    } catch (err) {
      // file:// history updates are optional.
    }
  }

  function scrollToAnchor(section, anchor) {
    var top = 0;
    if (anchor) {
      var target = document.getElementById(section.getAttribute("data-doc-id") + "-" + anchor);
      if (target) {
        top = target.getBoundingClientRect().top + window.pageYOffset - 72;
      }
    }
    var html = document.documentElement;
    var previous = html.style.scrollBehavior;
    html.style.scrollBehavior = "auto";
    window.scrollTo(0, top);
    html.style.scrollBehavior = previous;
  }

  function showDoc(docID, anchor, updateHash) {
    if (sections.length === 0) {
      return;
    }
    var section = null;
    for (var i = 0; i < sections.length; i++) {
      if (sections[i].getAttribute("data-doc-id") === docID) {
        section = sections[i];
        break;
      }
    }
    if (!section) {
      section = sections[0];
    }

    for (var j = 0; j < sections.length; j++) {
      sections[j].hidden = sections[j] !== section;
    }
    for (var k = 0; k < docButtons.length; k++) {
      docButtons[k].classList.toggle(
        "active",
        docButtons[k].getAttribute("data-doc-target") === section.getAttribute("data-doc-id")
      );
    }

    var docTitle = section.getAttribute("data-title") || "OpenMD";
    if (titleEl) {
      titleEl.textContent = docTitle;
    }
    document.title = docTitle + " - OpenMD";

    replaceTOC(section);
    applyChromaTheme();

    scrollToAnchor(section, anchor);
    window.requestAnimationFrame(function () {
      scrollToAnchor(section, anchor);
    });

    if (updateHash !== false) {
      setHash(section.getAttribute("data-doc-id"), anchor || "");
    }
    collectHeadings(section);
  }

  function initialDoc() {
    var params = new URLSearchParams(window.location.hash.slice(1));
    var docID = params.get("doc");
    var anchor = params.get("anchor");
    if (docID) {
      showDoc(docID, anchor, false);
      return;
    }
    var section = currentSection();
    if (section) {
      showDoc(section.getAttribute("data-doc-id"), null, false);
    }
  }

  for (var i = 0; i < docButtons.length; i++) {
    docButtons[i].addEventListener("click", function () {
      showDoc(this.getAttribute("data-doc-target"), null, true);
      document.body.classList.remove("sidebar-open");
    });
  }

  for (var t = 0; t < tabButtons.length; t++) {
    tabButtons[t].addEventListener("click", function () {
      activateTab(this.getAttribute("data-tab"));
    });
  }

  document.addEventListener("click", function (event) {
    var target = event.target;
    if (!target || !target.closest) {
      return;
    }
    var link = target.closest('a[href^="#doc="]');
    if (!link) {
      return;
    }
    event.preventDefault();
    var params = new URLSearchParams(link.getAttribute("href").slice(1));
    showDoc(params.get("doc"), params.get("anchor"), true);
  });

  if (menuButton) {
    menuButton.addEventListener("click", function () {
      document.body.classList.toggle("sidebar-open");
    });
  }

  window.addEventListener("scroll", updateActiveTOC, { passive: true });
  window.addEventListener("resize", updateActiveTOC);

  if (themeQuery) {
    if (themeQuery.addEventListener) {
      themeQuery.addEventListener("change", applyChromaTheme);
    } else if (themeQuery.addListener) {
      themeQuery.addListener(applyChromaTheme);
    }
  }

  initialDoc();
  applyChromaTheme();

  var initialParams = new URLSearchParams(window.location.hash.slice(1));
  var initialAnchor = initialParams.get("anchor");
  if (initialAnchor) {
    window.addEventListener("load", function () {
      var section = currentSection();
      if (section) {
        scrollToAnchor(section, initialAnchor);
      }
    });
  }
})();
