(function () {
  const BASE = "";

  const homeEl = document.getElementById("home");
  const headerEl = document.getElementById("header");
  const resultsSection = document.getElementById("results-section");
  const resultsInfo = document.getElementById("results-info");
  const resultsDiv = document.getElementById("results");

  const queryInput = document.getElementById("query");
  const searchBtn = document.getElementById("search-btn");
  const datasetSelect = document.getElementById("dataset");
  const boostCheckbox = document.getElementById("boost");
  const limitSelect = document.getElementById("limit");

  const headerLogo = document.getElementById("header-logo");
  const headerQuery = document.getElementById("header-query");
  const headerBtn = document.getElementById("header-btn");
  const headerDataset = document.getElementById("header-dataset");
  const headerBoost = document.getElementById("header-boost");
  const headerLimit = document.getElementById("header-limit");

  const modalOverlay = document.getElementById("modal-overlay");
  const modalCloseBtn = document.getElementById("modal-close");
  const modalContent = document.getElementById("modal-content");

  let currentDataset = "fruitsA";

  function syncHeaderFromHome() {
    headerDataset.value = datasetSelect.value;
    headerBoost.checked = boostCheckbox.checked;
    headerLimit.value = limitSelect.value;
  }

  function syncHomeFromHeader() {
    datasetSelect.value = headerDataset.value;
    boostCheckbox.checked = headerBoost.checked;
    limitSelect.value = headerLimit.value;
  }

  function showResultsView(query) {
    homeEl.classList.add("hidden");
    headerEl.classList.remove("hidden");
    resultsSection.classList.remove("hidden");
    headerQuery.value = query;
    syncHeaderFromHome();
  }

  function showHomeView() {
    homeEl.classList.remove("hidden");
    headerEl.classList.add("hidden");
    resultsSection.classList.add("hidden");
    queryInput.focus();
  }

  async function performSearch(source) {
    const isHeader = source === "header";

    if (isHeader) {
      syncHomeFromHeader();
    } else {
      syncHeaderFromHome();
    }

    const q = (isHeader ? headerQuery.value : queryInput.value).trim();
    const dataset = isHeader ? headerDataset.value : datasetSelect.value;
    const boost = isHeader ? headerBoost.checked : boostCheckbox.checked;
    const limit = isHeader ? headerLimit.value : limitSelect.value;

    if (!q) {
      return;
    }

    currentDataset = dataset;
    showResultsView(q);

    const params = new URLSearchParams({
      q,
      boost: String(boost),
      limit: String(limit),
    });

    const url = `${BASE}/${dataset}?${params.toString()}`;

    resultsInfo.textContent = "";
    resultsDiv.innerHTML = '<div class="loading">Searching</div>';

    try {
      const response = await fetch(url);

      if (!response.ok) {
        throw new Error(`Server returned ${response.status}`);
      }

      const data = await response.json();
      const items = Array.isArray(data.result) ? data.result : [];

      resultsInfo.textContent = `About ${items.length} result${items.length !== 1 ? "s" : ""} from "${dataset}" dataset`;
      renderResults(items);
    } catch (error) {
      resultsInfo.textContent = "";
      resultsDiv.innerHTML = `<div class="no-results">Error: ${escapeHtml(error.message)}</div>`;
    }
  }

  function renderResults(items) {
    if (!items.length) {
      resultsDiv.innerHTML = '<div class="no-results">No results found. Try a different query.</div>';
      return;
    }

    resultsDiv.innerHTML = items
      .map((item) => {
        const title = item.title || item.url || "Untitled";
        const url = item.url || "#";
        const snippet = item.snippet || "";
        const score =
          typeof item.score === "number" ? item.score.toFixed(4) : "—";
        const pagerank =
          typeof item.pr === "number" ? item.pr.toFixed(6) : "—";

        return `
          <article class="result">
            <a class="result-title" href="${escapeAttr(url)}" target="_blank" rel="noopener noreferrer">
              ${escapeHtml(title)}
            </a>
            <div class="result-url">${escapeHtml(url)}</div>
            <div class="result-snippet">${escapeHtml(snippet)}</div>
            <div class="result-meta">
              <span>Score: ${score}</span>
              <span>PageRank: ${pagerank}</span>
              <span>Dataset: ${escapeHtml(currentDataset)}</span>
            </div>
            <div class="result-actions">
              <button class="details-btn" type="button" data-url="${escapeAttr(url)}">
                Page details
              </button>
            </div>
          </article>
        `;
      })
      .join("");
  }

  async function showPageDetails(url) {
    modalOverlay.classList.remove("hidden");
    modalContent.innerHTML = '<div class="loading">Loading details</div>';

    const params = new URLSearchParams({
      dataset: currentDataset,
      url: url,
    });

    try {
      const response = await fetch(`${BASE}/page?${params.toString()}`);

      if (!response.ok) {
        throw new Error(`Server returned ${response.status}`);
      }

      const data = await response.json();
      modalContent.innerHTML = buildDetailsHTML(data);
    } catch (error) {
      modalContent.innerHTML = `<p style="color:#d93025;">Error: ${escapeHtml(error.message)}</p>`;
    }
  }

  function buildDetailsHTML(data) {
    const incoming = Array.isArray(data.incoming_links) ? data.incoming_links : [];
    const outgoing = Array.isArray(data.outgoing_links) ? data.outgoing_links : [];
    const words = data.word_counts && typeof data.word_counts === "object" ? data.word_counts : {};

    const wordEntries = Object.entries(words).sort((a, b) => b[1] - a[1]).slice(0, 30);

    return `
      <h2 id="modal-title">${escapeHtml(data.title || data.url || "Page Details")}</h2>

      <div class="detail-row">
        <div class="detail-label">URL</div>
        <div class="detail-value">
          <a href="${escapeAttr(data.url || "#")}" target="_blank" rel="noopener noreferrer">
            ${escapeHtml(data.url || "—")}
          </a>
        </div>
      </div>

      <div class="detail-row">
        <div class="detail-label">Dataset</div>
        <div class="detail-value">${escapeHtml(data.dataset || currentDataset)}</div>
      </div>

      <div class="detail-row">
        <div class="detail-label">PageRank</div>
        <div class="detail-value">${typeof data.pr === "number" ? data.pr.toFixed(6) : "—"}</div>
      </div>

      <div class="detail-row">
        <div class="detail-label">Incoming Links (${incoming.length})</div>
        <div class="detail-value">
          ${
            incoming.length
              ? `<ul class="link-list">
                  ${incoming
                    .map(
                      (link) => `
                        <li>
                          <a href="${escapeAttr(link)}" target="_blank" rel="noopener noreferrer">
                            ${escapeHtml(link)}
                          </a>
                        </li>
                      `
                    )
                    .join("")}
                </ul>`
              : '<span style="color:#70757a;">None</span>'
          }
        </div>
      </div>

      <div class="detail-row">
        <div class="detail-label">Outgoing Links (${outgoing.length})</div>
        <div class="detail-value">
          ${
            outgoing.length
              ? `<ul class="link-list">
                  ${outgoing
                    .map(
                      (link) => `
                        <li>
                          <a href="${escapeAttr(link)}" target="_blank" rel="noopener noreferrer">
                            ${escapeHtml(link)}
                          </a>
                        </li>
                      `
                    )
                    .join("")}
                </ul>`
              : '<span style="color:#70757a;">None</span>'
          }
        </div>
      </div>

      ${
        wordEntries.length
          ? `
            <div class="detail-row">
              <div class="detail-label">Word Counts (Top ${wordEntries.length})</div>
              <div class="detail-value">
                <div class="word-counts">
                  ${wordEntries
                    .map(
                      ([word, count]) =>
                        `<span class="word-chip">${escapeHtml(word)} × ${escapeHtml(String(count))}</span>`
                    )
                    .join("")}
                </div>
              </div>
            </div>
          `
          : ""
      }
    `;
  }

  function closeModal() {
    modalOverlay.classList.add("hidden");
    modalContent.innerHTML = "";
  }

  function escapeHtml(value) {
    const div = document.createElement("div");
    div.textContent = value == null ? "" : String(value);
    return div.innerHTML;
  }

  function escapeAttr(value) {
    return String(value == null ? "" : value)
      .replace(/&/g, "&amp;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
  }

  searchBtn.addEventListener("click", function () {
    performSearch("home");
  });

  queryInput.addEventListener("keydown", function (event) {
    if (event.key === "Enter") {
      performSearch("home");
    }
  });

  headerBtn.addEventListener("click", function () {
    performSearch("header");
  });

  headerQuery.addEventListener("keydown", function (event) {
    if (event.key === "Enter") {
      performSearch("header");
    }
  });

  headerLogo.addEventListener("click", function () {
    queryInput.value = headerQuery.value;
    syncHomeFromHeader();
    showHomeView();
  });

  resultsDiv.addEventListener("click", function (event) {
    const button = event.target.closest(".details-btn");
    if (!button) return;

    const url = button.getAttribute("data-url");
    if (url) {
      showPageDetails(url);
    }
  });

  modalCloseBtn.addEventListener("click", closeModal);

  modalOverlay.addEventListener("click", function (event) {
    if (event.target === modalOverlay) {
      closeModal();
    }
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape" && !modalOverlay.classList.contains("hidden")) {
      closeModal();
    }
  });

  queryInput.focus();
})();