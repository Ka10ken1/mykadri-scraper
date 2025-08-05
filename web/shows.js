let shows = [];
let allShows = [];
let currentPage = 1;
let highlightedIndex = -1;
const pageSize = 50;

async function fetchShows() {
  try {
    const res = await fetch("/api/shows/images");
    if (!res.ok) throw new Error("Failed to fetch shows");
    shows = await res.json();
    allShows = [...shows];
    currentPage = 1;
    updateStatusBar();
    renderPage();
  } catch (err) {
    console.error("Error loading shows:", err);
  }
}

function updateStatusBar(status = "READY") {
  const statusBar = document.querySelector(".status-bar");
  const count = shows.length;
  const maxPage = Math.ceil(count / pageSize);
  statusBar.innerHTML = `
      <span>STATUS: ${status}</span>
      <span>SHOWS: ${count}</span>
      <span>PAGES: ${maxPage}</span>
  `;
}

function renderPage() {
  const start = (currentPage - 1) * pageSize;
  const end = start + pageSize;
  const currentItems = shows.slice(start, end);
  const grid = document.getElementById("movie-grid");
  grid.innerHTML = "";

  currentItems.forEach((show) => {
    const item = document.createElement("div");
    item.className = "movie-item";

    const link = document.createElement("a");
    link.href = `/api/show/${show.id}`;
    link.style.textDecoration = "none";
    link.style.color = "inherit";

    const img = document.createElement("img");
    img.src = show.image;
    img.alt = show.Title || show.title || show.id;
    img.className = "movie-poster";

    const title = document.createElement("div");
    title.className = "movie-title";
    title.textContent = show.Title || show.title || "Untitled";

    const englishTitle = document.createElement("div");
    englishTitle.className = "movie-english-title";
    englishTitle.textContent = show.TitleEnglish || show.titleEnglish || "";
    if (englishTitle.textContent) {
      title.style.fontSize = "11px";
      title.style.marginBottom = "2px";
      englishTitle.style.color = "#888888";
      englishTitle.style.fontSize = "10px";
      englishTitle.style.marginBottom = "4px";
    }

    const id = document.createElement("div");
    id.className = "movie-id";
    id.textContent = `ID: ${show.id}`;

    link.appendChild(img);
    item.appendChild(link);
    item.appendChild(title);
    item.appendChild(englishTitle);
    item.appendChild(id);
    grid.appendChild(item);

    item.addEventListener("click", (e) => {
      if (e.target.tagName !== "A" && e.target.tagName !== "IMG") {
        document
          .querySelectorAll(".movie-item")
          .forEach((i) => i.classList.remove("selected"));
        item.classList.add("selected");
      }
    });
  });

  document.getElementById("page-number").textContent = `PAGE ${currentPage}`;
  updateNavigationButtons();
}

function updateNavigationButtons() {
  const maxPage = Math.ceil(shows.length / pageSize);
  document.getElementById("prev").disabled = currentPage === 1;
  document.getElementById("next").disabled = currentPage >= maxPage;
}

function showSearchResults(results, query) {
  const searchResults = document.getElementById("searchResults");
  if (!query || query.length < 2) {
    hideSearchResults();
    return;
  }
  if (results.length === 0) {
    searchResults.innerHTML = '<div class="no-results">No shows found</div>';
    searchResults.classList.add("show");
    return;
  }
  const limited = results.slice(0, 8);
  searchResults.innerHTML = limited
    .map(
      (show, idx) => `
    <div class="search-result-item" data-index="${idx}" data-show-id="${show.id}">
      <img src="${show.image}" alt="${show.Title || show.title}" class="search-result-poster" />
      <div class="search-result-info">
        <div class="search-result-title">${show.Title || show.title}
          ${show.TitleEnglish ? `<span style="color:#666; font-size:10px;">(${show.TitleEnglish})</span>` : ""}
        </div>
        <div class="search-result-id">ID: ${show.id}</div>
      </div>
    </div>`,
    )
    .join("");
  searchResults.classList.add("show");
  highlightedIndex = -1;
}

function hideSearchResults() {
  const searchResults = document.getElementById("searchResults");
  searchResults.classList.remove("show");
  highlightedIndex = -1;
}

function highlightResult(index) {
  const items = document.querySelectorAll(".search-result-item");
  items.forEach((item, i) => item.classList.toggle("highlighted", i === index));
  highlightedIndex = index;
}

function selectSearchResult(showId) {
  const show = allShows.find((m) => m.id === showId);
  if (show) {
    shows = [show];
    currentPage = 1;
    updateStatusBar("SELECTED");
    renderPage();
    document.getElementById("searchInput").value = show.Title || show.title;
    hideSearchResults();
  }
}

async function performSearch(query) {
  if (!query || query.length < 2) {
    shows = [...allShows];
    currentPage = 1;
    updateStatusBar("READY");
    renderPage();
    hideSearchResults();
    return;
  }

  const localResults = allShows.filter((show) => {
    const title = (show.Title || show.title || "").toLowerCase();
    const english = (show.TitleEnglish || "").toLowerCase();
    const id = show.id.toLowerCase();
    const q = query.toLowerCase();
    return title.includes(q) || english.includes(q) || id.includes(q);
  });

  showSearchResults(localResults, query);

  try {
    const res = await fetch(`/api/search-shows?q=${encodeURIComponent(query)}`);
    if (res.ok) {
      let apiResults = await res.json();
      apiResults = apiResults.map((m) => ({
        id: m._id || m.id,
        Title: m.title || m.Title || "",
        TitleEnglish: m.titleEnglish || m.TitleEnglish || "",
        image: m.image || m.Image || "",
      }));
      shows = apiResults;
      currentPage = 1;
      updateStatusBar("SEARCH (API)");
      renderPage();
    } else {
      shows = localResults;
      currentPage = 1;
      updateStatusBar("SEARCH (local fallback)");
      renderPage();
    }
  } catch (err) {
    console.error("API search error:", err);
    shows = localResults;
    currentPage = 1;
    updateStatusBar("SEARCH (local fallback)");
    renderPage();
  }
}

function debounce(fn, delay) {
  let timeout;
  return (...args) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => fn(...args), delay);
  };
}

const debouncedSearch = debounce(performSearch, 200);

document.getElementById("prev").addEventListener("click", () => {
  if (currentPage > 1) {
    currentPage--;
    renderPage();
  }
});
document.getElementById("next").addEventListener("click", () => {
  if (currentPage < Math.ceil(shows.length / pageSize)) {
    currentPage++;
    renderPage();
  }
});
const searchInput = document.getElementById("searchInput");
searchInput.addEventListener("input", (e) => debouncedSearch(e.target.value));
searchInput.addEventListener("keydown", (e) => {
  const items = document.querySelectorAll(".search-result-item");
  switch (e.key) {
    case "ArrowDown":
      e.preventDefault();
      if (items.length) {
        highlightedIndex = Math.min(highlightedIndex + 1, items.length - 1);
        highlightResult(highlightedIndex);
      }
      break;
    case "ArrowUp":
      e.preventDefault();
      if (items.length) {
        highlightedIndex = Math.max(highlightedIndex - 1, -1);
        highlightResult(highlightedIndex);
      }
      break;
    case "Enter":
      e.preventDefault();
      if (highlightedIndex >= 0 && items[highlightedIndex]) {
        selectSearchResult(items[highlightedIndex].dataset.showId);
      }
      break;
    case "Escape":
      hideSearchResults();
      searchInput.blur();
      break;
  }
});
document.addEventListener("click", (e) => {
  if (e.target.closest(".search-result-item")) {
    const showId = e.target.closest(".search-result-item").dataset.showId;
    selectSearchResult(showId);
  } else if (!e.target.closest(".search-bar")) {
    hideSearchResults();
  }
});

fetchShows();
