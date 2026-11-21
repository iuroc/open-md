(function () {
  "use strict";

  var sections = Array.prototype.slice.call(document.querySelectorAll(".doc-section"));
  var docButtons = Array.prototype.slice.call(document.querySelectorAll("[data-doc-target]"));
  var treeToggles = Array.prototype.slice.call(document.querySelectorAll("[data-tree-folder]"));
  var docList = document.getElementById("doc-list");
  var isDocPage = document.body.getAttribute("data-doc-page") === "true";
  var rootKey = document.body.getAttribute("data-root-key") || "";
  var activePath = document.body.getAttribute("data-active-path") || "";
  var treeStorageKey = "openmd-tree:" + rootKey;
  var sourceStorageKey = "openmd-source:" + rootKey;
  var tabButtons = Array.prototype.slice.call(document.querySelectorAll(".sidebar-tab"));
  var tabPanels = Array.prototype.slice.call(document.querySelectorAll(".tab-panel"));
  var toc = document.getElementById("toc");
  var titleEl = document.getElementById("doc-title");
  var menuButton = document.getElementById("menu-toggle");
  var widthButton = document.getElementById("width-toggle");
  var sourceButton = document.getElementById("source-toggle");
  var sidebarBackdrop = document.getElementById("sidebar-backdrop");
  var sidebarClose = document.getElementById("sidebar-close");
  var settingsButton = document.getElementById("settings-button");
  var settingsPanel = document.getElementById("settings-panel");
  var settingsClose = document.getElementById("settings-close");
  var settingsTabButtons = Array.prototype.slice.call(document.querySelectorAll(".settings-tab"));
  var settingsTabPanels = Array.prototype.slice.call(document.querySelectorAll(".settings-tabpanel"));
  var settingsTabsEl = document.querySelector(".settings-tabs");
  var filePathEl = document.getElementById("file-path");
  var fileDirEl = document.getElementById("file-dir");
  var copyPathButton = document.getElementById("copy-path");
  var copyStatusEl = document.getElementById("copy-status");
  var currentPath = "";
  var copyStatusTimer = null;
  var copyGeneration = 0;
  var themeButtons = Array.prototype.slice.call(document.querySelectorAll(".theme-option"));
  var validThemes = ["blue", "teal", "violet", "graphite"];
  var themeQuery = window.matchMedia ? window.matchMedia("(prefers-color-scheme: dark)") : null;
  var mobileQuery = window.matchMedia ? window.matchMedia("(max-width: 900px)") : null;
  var activeHeadings = [];

  function applyChromaTheme() {
    var dark = !!(themeQuery && themeQuery.matches);
    var blocks = document.querySelectorAll(".chroma");
    for (var i = 0; i < blocks.length; i++) {
      blocks[i].classList.toggle("dark", dark);
      blocks[i].classList.toggle("light", !dark);
    }
  }

  function readPref(name) {
    try {
      return window.localStorage.getItem(name);
    } catch (err) {
      return null;
    }
  }

  function writePref(name, value) {
    try {
      window.localStorage.setItem(name, value);
    } catch (err) {
      // Preferences are optional for file:// pages.
    }
  }

  function isMobile() {
    return mobileQuery ? mobileQuery.matches : window.innerWidth <= 900;
  }

  function sidebarVisible() {
    if (isMobile()) {
      return document.body.classList.contains("sidebar-open");
    }
    return !document.body.classList.contains("sidebar-collapsed");
  }

  function syncMenuButton() {
    if (!menuButton) {
      return;
    }
    var visible = sidebarVisible();
    menuButton.setAttribute("aria-expanded", visible ? "true" : "false");
    var label = visible ? "收起侧栏" : "展开侧栏";
    menuButton.setAttribute("aria-label", label);
    menuButton.title = label;
  }

  function syncWidthButton() {
    if (!widthButton) {
      return;
    }
    var wide = document.body.classList.contains("full-width");
    widthButton.setAttribute("aria-pressed", wide ? "true" : "false");
    var label = wide ? "切换为固定宽度" : "切换全宽阅读";
    widthButton.setAttribute("aria-label", label);
    widthButton.title = label;
  }

  function setSettingsOpen(open) {
    if (!settingsPanel || !settingsButton) {
      return;
    }
    settingsPanel.hidden = !open;
    settingsButton.setAttribute("aria-expanded", open ? "true" : "false");
  }

  function activateSettingsTab(name) {
    if (settingsTabButtons.length === 0) {
      return;
    }
    for (var i = 0; i < settingsTabButtons.length; i++) {
      var selected = settingsTabButtons[i].getAttribute("data-settings-tab") === name;
      settingsTabButtons[i].classList.toggle("active", selected);
      settingsTabButtons[i].setAttribute("aria-selected", selected ? "true" : "false");
      settingsTabButtons[i].tabIndex = selected ? 0 : -1;
    }
    for (var j = 0; j < settingsTabPanels.length; j++) {
      settingsTabPanels[j].hidden =
        settingsTabPanels[j].getAttribute("data-settings-panel") !== name;
    }
  }

  function sourceModeActive() {
    return document.body.classList.contains("source-mode");
  }

  function syncSourceButton() {
    if (!sourceButton) {
      return;
    }
    var active = sourceModeActive();
    sourceButton.setAttribute("aria-pressed", active ? "true" : "false");
    var label = active ? "返回阅读视图" : "查看源码";
    sourceButton.setAttribute("aria-label", label);
    sourceButton.title = label;
  }

  function setSourceMode(active) {
    if (sourceModeActive() === active) {
      syncSourceButton();
      return;
    }
    var y = window.scrollY;
    document.body.classList.toggle("source-mode", active);
    syncSourceButton();
    var max = document.documentElement.scrollHeight - window.innerHeight;
    window.scrollTo(0, Math.min(y, Math.max(0, max)));
  }

  function currentTheme() {
    var theme = document.documentElement.getAttribute("data-theme");
    return validThemes.indexOf(theme) >= 0 ? theme : "blue";
  }

  function applyTheme(theme, persist) {
    if (validThemes.indexOf(theme) < 0) {
      theme = "blue";
    }
    document.documentElement.setAttribute("data-theme", theme);
    for (var i = 0; i < themeButtons.length; i++) {
      themeButtons[i].setAttribute(
        "aria-pressed",
        themeButtons[i].getAttribute("data-theme") === theme ? "true" : "false"
      );
    }
    if (persist) {
      writePref("openmd-theme", theme);
    }
  }

  function setCopyStatus(text) {
    if (copyStatusTimer) {
      window.clearTimeout(copyStatusTimer);
      copyStatusTimer = null;
    }
    if (copyStatusEl) {
      copyStatusEl.textContent = text;
    }
    if (text) {
      copyStatusTimer = window.setTimeout(function () {
        copyStatusTimer = null;
        if (copyStatusEl) {
          copyStatusEl.textContent = "";
        }
      }, 2000);
    }
  }

  function fallbackCopy(text) {
    var previousFocus = document.activeElement;
    var area = document.createElement("textarea");
    area.value = text;
    area.setAttribute("readonly", "");
    area.style.position = "fixed";
    area.style.opacity = "0";
    document.body.appendChild(area);
    area.select();
    var ok = false;
    try {
      ok = document.execCommand("copy");
    } catch (err) {
      ok = false;
    }
    document.body.removeChild(area);
    if (previousFocus && previousFocus.focus) {
      previousFocus.focus();
    }
    return ok;
  }

  function writeClipboard(text, done) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      try {
        navigator.clipboard.writeText(text).then(
          function () {
            done(true);
          },
          function () {
            done(fallbackCopy(text));
          }
        );
        return;
      } catch (err) {
        // Fall through to the textarea fallback.
      }
    }
    done(fallbackCopy(text));
  }

  function copyCurrentPath() {
    if (!currentPath) {
      setCopyStatus("没有可复制的路径");
      return;
    }
    var request = ++copyGeneration;
    var requestedPath = currentPath;
    writeClipboard(requestedPath, function (ok) {
      if (request !== copyGeneration || currentPath !== requestedPath) {
        return;
      }
      setCopyStatus(ok ? "已复制" : "复制失败");
    });
  }

  function setSourceStatus(el, text) {
    if (!el) {
      return;
    }
    if (el._sourceStatusTimer) {
      window.clearTimeout(el._sourceStatusTimer);
      el._sourceStatusTimer = null;
    }
    el.textContent = text;
    if (text) {
      el._sourceStatusTimer = window.setTimeout(function () {
        el._sourceStatusTimer = null;
        el.textContent = "";
      }, 2000);
    }
  }

  function copySource() {
    var section = currentSection();
    var code = section ? section.querySelector(".source-code") : null;
    var status = section ? section.querySelector(".source-copy-status") : null;
    var text = code ? code.textContent : "";
    if (!text) {
      setSourceStatus(status, "没有可复制的源码");
      return;
    }
    writeClipboard(text, function (ok) {
      setSourceStatus(status, ok ? "已复制" : "复制失败");
    });
  }

  function closeMobileSidebar() {
    document.body.classList.remove("sidebar-open");
    syncMenuButton();
  }

  function onViewportChange() {
    if (!isMobile()) {
      document.body.classList.remove("sidebar-open");
    }
    syncMenuButton();
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
    if (!toc || activeHeadings.length === 0 || sourceModeActive()) {
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
      var target =
        document.getElementById(anchor) ||
        document.getElementById(section.getAttribute("data-doc-id") + "-" + anchor);
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
      var isActive = docButtons[k].getAttribute("data-doc-target") === section.getAttribute("data-doc-id");
      docButtons[k].classList.toggle("active", isActive);
      if (docButtons[k].classList.contains("tree-file-button")) {
        if (isActive) {
          docButtons[k].setAttribute("aria-current", "true");
        } else {
          docButtons[k].removeAttribute("aria-current");
        }
      }
    }
    revealTreeTarget(section.getAttribute("data-doc-id"));

    var docTitle = section.getAttribute("data-title") || "OpenMD";
    if (titleEl) {
      titleEl.textContent = docTitle;
    }
    document.title = docTitle + " - OpenMD";

    currentPath = section.getAttribute("data-path") || "";
    if (filePathEl) {
      filePathEl.textContent = currentPath;
    }
    if (fileDirEl) {
      fileDirEl.textContent = section.getAttribute("data-dir") || "";
    }
    copyGeneration += 1;
    setCopyStatus("");

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

  function revealTreeTarget(docID) {
    var target = null;
    for (var i = 0; i < docButtons.length; i++) {
      if (
        docButtons[i].getAttribute("data-doc-target") === docID &&
        docButtons[i].classList.contains("tree-file-button")
      ) {
        target = docButtons[i];
        break;
      }
    }
    if (!target) {
      return;
    }
    var node = target.parentElement;
    while (node && node.classList) {
      if (node.classList.contains("tree-children")) {
        node.hidden = false;
        var folder = node.parentElement;
        var toggle = folder ? folder.querySelector(".tree-toggle") : null;
        if (toggle) {
          toggle.setAttribute("aria-expanded", "true");
        }
      }
      node = node.parentElement;
    }
    if (target.scrollIntoView) {
      target.scrollIntoView({ block: "nearest" });
    }
  }

  for (var i = 0; i < treeToggles.length; i++) {
    treeToggles[i].addEventListener("click", function () {
      setFolderExpanded(this, this.getAttribute("aria-expanded") !== "true", true);
    });
  }

  function readTreeExpanded() {
    try {
      return JSON.parse(sessionStorage.getItem(treeStorageKey) || "[]");
    } catch (err) {
      return [];
    }
  }

  function saveTreeExpanded() {
    if (!docList) {
      return;
    }
    var paths = [];
    var toggles = docList.querySelectorAll(".tree-toggle");
    for (var i = 0; i < toggles.length; i++) {
      if (toggles[i].getAttribute("aria-expanded") === "true") {
        paths.push(toggles[i].getAttribute("data-tree-path"));
      }
    }
    try {
      sessionStorage.setItem(treeStorageKey, JSON.stringify(paths));
    } catch (err) {}
  }

  function setFolderExpanded(toggle, expanded, persist) {
    toggle.setAttribute("aria-expanded", expanded ? "true" : "false");
    var folder = toggle.parentElement;
    var children = folder ? folder.querySelector(".tree-children") : null;
    if (children) {
      children.hidden = !expanded;
    }
    if (persist) {
      saveTreeExpanded();
    }
  }

  function expandTreeAncestors(element) {
    var node = element.parentElement;
    while (node && node !== docList) {
      if (node.classList && node.classList.contains("tree-children")) {
        node.hidden = false;
        var folder = node.parentElement;
        var toggle = folder ? folder.querySelector(".tree-toggle") : null;
        if (toggle) {
          toggle.setAttribute("aria-expanded", "true");
        }
      }
      node = node.parentElement;
    }
  }

  function markActiveTree() {
    if (!docList || !activePath) {
      return;
    }
    var links = docList.querySelectorAll(".tree-file-link, .doc-item[data-tree-path]");
    for (var i = 0; i < links.length; i++) {
      if (links[i].getAttribute("data-tree-path") !== activePath) {
        continue;
      }
      links[i].classList.add("active");
      links[i].setAttribute("aria-current", "page");
      expandTreeAncestors(links[i]);
      if (links[i].scrollIntoView) {
        links[i].scrollIntoView({ block: "nearest" });
      }
      return;
    }
  }

  function initTree() {
    if (!docList || !window.OPENMD_TREE_HTML) {
      return;
    }
    docList.innerHTML = window.OPENMD_TREE_HTML;
    var stored = readTreeExpanded();
    var toggles = docList.querySelectorAll(".tree-toggle");
    for (var i = 0; i < toggles.length; i++) {
      if (stored.indexOf(toggles[i].getAttribute("data-tree-path")) >= 0) {
        setFolderExpanded(toggles[i], true, false);
      }
    }
    markActiveTree();
    saveTreeExpanded();
    docList.addEventListener("click", function (event) {
      var target = event.target;
      if (!target || !target.closest) {
        return;
      }
      var toggle = target.closest(".tree-toggle");
      if (toggle && docList.contains(toggle)) {
        setFolderExpanded(toggle, toggle.getAttribute("aria-expanded") !== "true", true);
      }
    });
  }

  function navigationType() {
    try {
      var entries = performance.getEntriesByType("navigation");
      return entries.length ? entries[0].type : "";
    } catch (err) {
      return "";
    }
  }

  function persistSourceMode(active) {
    try {
      if (active) {
        sessionStorage.setItem(sourceStorageKey, "1");
      } else {
        sessionStorage.removeItem(sourceStorageKey);
      }
    } catch (err) {}
  }

  function restoreSourceMode() {
    if (!isDocPage) {
      return;
    }
    if (navigationType() === "reload") {
      persistSourceMode(false);
      return;
    }
    try {
      if (sessionStorage.getItem(sourceStorageKey) === "1") {
        setSourceMode(true);
      }
    } catch (err) {}
  }

  function initDocPage() {
    var section = sections.length ? sections[0] : null;
    if (!section) {
      return;
    }
    var anchor = "";
    if (window.location.hash && window.location.hash.indexOf("#doc=") !== 0) {
      try {
        anchor = decodeURIComponent(window.location.hash.slice(1));
      } catch (err) {
        anchor = window.location.hash.slice(1);
      }
    }
    showDoc(section.getAttribute("data-doc-id"), anchor, false);
    initTree();
    restoreSourceMode();
  }

  for (var i = 0; i < docButtons.length; i++) {
    docButtons[i].addEventListener("click", function () {
      showDoc(this.getAttribute("data-doc-target"), null, true);
      closeMobileSidebar();
    });
  }

  for (var t = 0; t < tabButtons.length; t++) {
    tabButtons[t].addEventListener("click", function () {
      activateTab(this.getAttribute("data-tab"));
    });
  }

  for (var s = 0; s < settingsTabButtons.length; s++) {
    settingsTabButtons[s].addEventListener("click", function () {
      activateSettingsTab(this.getAttribute("data-settings-tab"));
    });
  }

  if (settingsTabsEl) {
    settingsTabsEl.addEventListener("keydown", function (event) {
      var index = settingsTabButtons.indexOf(document.activeElement);
      if (index < 0) {
        return;
      }
      var next = index;
      if (event.key === "ArrowRight") {
        next = (index + 1) % settingsTabButtons.length;
      } else if (event.key === "ArrowLeft") {
        next = (index - 1 + settingsTabButtons.length) % settingsTabButtons.length;
      } else if (event.key === "Home") {
        next = 0;
      } else if (event.key === "End") {
        next = settingsTabButtons.length - 1;
      } else {
        return;
      }
      event.preventDefault();
      activateSettingsTab(settingsTabButtons[next].getAttribute("data-settings-tab"));
      settingsTabButtons[next].focus();
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
      if (isMobile()) {
        document.body.classList.toggle("sidebar-open");
      } else {
        document.body.classList.toggle("sidebar-collapsed");
        writePref(
          "openmd-sidebar",
          document.body.classList.contains("sidebar-collapsed") ? "collapsed" : "open"
        );
      }
      syncMenuButton();
    });
  }

  if (sidebarBackdrop) {
    sidebarBackdrop.addEventListener("click", closeMobileSidebar);
  }

  if (sidebarClose) {
    sidebarClose.addEventListener("click", closeMobileSidebar);
  }

  if (widthButton) {
    widthButton.addEventListener("click", function () {
      document.body.classList.toggle("full-width");
      syncWidthButton();
      writePref(
        "openmd-width",
        document.body.classList.contains("full-width") ? "wide" : "fixed"
      );
    });
  }

  if (sourceButton) {
    sourceButton.addEventListener("click", function () {
      var next = !sourceModeActive();
      setSourceMode(next);
      persistSourceMode(next);
    });
  }

  document.addEventListener("click", function (event) {
    var target = event.target;
    if (!target || !target.closest) {
      return;
    }
    if (target.closest(".source-copy")) {
      copySource();
    }
  });

  if (toc) {
    toc.addEventListener("click", function (event) {
      var target = event.target;
      if (!target || !target.closest || !target.closest(".toc-link")) {
        return;
      }
      if (sourceModeActive()) {
        document.body.classList.remove("source-mode");
        syncSourceButton();
        persistSourceMode(false);
      }
    });
  }

  if (settingsButton) {
    settingsButton.addEventListener("click", function (event) {
      event.stopPropagation();
      setSettingsOpen(settingsPanel ? settingsPanel.hidden : false);
    });
  }

  if (settingsClose) {
    settingsClose.addEventListener("click", function () {
      setSettingsOpen(false);
    });
  }

  if (copyPathButton) {
    copyPathButton.addEventListener("click", copyCurrentPath);
  }

  for (var th = 0; th < themeButtons.length; th++) {
    themeButtons[th].addEventListener("click", function () {
      applyTheme(this.getAttribute("data-theme"), true);
    });
  }

  if (settingsPanel) {
    settingsPanel.addEventListener("click", function (event) {
      var target = event.target;
      if (target && target.closest && target.closest("a")) {
        setSettingsOpen(false);
      }
    });
  }

  document.addEventListener("click", function (event) {
    if (!settingsPanel || settingsPanel.hidden) {
      return;
    }
    var target = event.target;
    if (settingsPanel.contains(target) || (settingsButton && settingsButton.contains(target))) {
      return;
    }
    setSettingsOpen(false);
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape") {
      setSettingsOpen(false);
      closeMobileSidebar();
    }
  });

  document.addEventListener("keydown", function (event) {
    if (!(event.ctrlKey || event.metaKey) || event.shiftKey || event.altKey) {
      return;
    }
    if (event.key !== "a" && event.key !== "A") {
      return;
    }
    var target = event.target;
    var code = target && target.closest ? target.closest(".source-code") : null;
    if (!code) {
      return;
    }
    event.preventDefault();
    var range = document.createRange();
    range.selectNodeContents(code);
    var selection = window.getSelection();
    selection.removeAllRanges();
    selection.addRange(range);
  });

  if (mobileQuery) {
    if (mobileQuery.addEventListener) {
      mobileQuery.addEventListener("change", onViewportChange);
    } else if (mobileQuery.addListener) {
      mobileQuery.addListener(onViewportChange);
    }
  }

  window.addEventListener("scroll", updateActiveTOC, { passive: true });
  window.addEventListener("resize", updateActiveTOC);
  window.addEventListener("resize", onViewportChange);

  if (themeQuery) {
    if (themeQuery.addEventListener) {
      themeQuery.addEventListener("change", applyChromaTheme);
    } else if (themeQuery.addListener) {
      themeQuery.addListener(applyChromaTheme);
    }
  }

  if (readPref("openmd-width") === "wide") {
    document.body.classList.add("full-width");
  }
  if (readPref("openmd-sidebar") === "collapsed") {
    document.body.classList.add("sidebar-collapsed");
  }
  syncWidthButton();
  syncMenuButton();
  syncSourceButton();
  setSettingsOpen(false);

  var storedTheme = readPref("openmd-theme");
  applyTheme(validThemes.indexOf(storedTheme) >= 0 ? storedTheme : currentTheme(), false);

  if (isDocPage) {
    initDocPage();
  } else {
    initialDoc();
  }
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
