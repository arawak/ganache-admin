document.addEventListener("DOMContentLoaded", () => {
  setupPasteUpload();
  setupCopyButtons();
});

function setupPasteUpload() {
  const pasteZone = document.getElementById("paste-zone");
  const preview = document.getElementById("paste-preview");
  const form = document.getElementById("upload-form");
  if (!pasteZone || !form) return;

  let pastedFile = null;

  pasteZone.addEventListener("paste", (event) => {
    const items = Array.from(event.clipboardData?.items || []);
    const imageItem = items.find((i) => i.type && i.type.startsWith("image"));
    if (!imageItem) return;
    const file = imageItem.getAsFile();
    if (!file) return;
    pastedFile = new File([file], file.name || "pasted.png", { type: file.type });
    if (preview) {
      const url = URL.createObjectURL(file);
      preview.innerHTML = `<img src="${url}" style="max-width:240px; border-radius:8px;">`;
    }
  });

  form.addEventListener("submit", async (event) => {
    if (!pastedFile) return;
    event.preventDefault();
    const fd = new FormData(form);
    fd.set("file", pastedFile, pastedFile.name);
    const resp = await fetch(form.action, { method: "POST", body: fd, redirect: "follow" });
    if (resp.redirected) {
      window.location = resp.url;
      return;
    }
    if (resp.ok) {
      const text = await resp.text();
      document.open();
      document.write(text);
      document.close();
    }
  });
}

function setupCopyButtons() {
  document.querySelectorAll("[data-copy]").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const value = btn.getAttribute("data-copy");
      if (!value) return;
      try {
        await navigator.clipboard.writeText(value);
        btn.textContent = "Copied";
        setTimeout(() => (btn.textContent = "Copy"), 1200);
      } catch (err) {
        console.error("copy failed", err);
      }
    });
  });
}
