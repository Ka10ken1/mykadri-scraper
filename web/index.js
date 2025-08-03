let movies = [];
let currentPage = 1;
const pageSize = 52;

async function fetchMovies() {
  try {
    const res = await fetch("/movie-images");
    if (!res.ok) throw new Error("Failed to fetch movies");
    movies = await res.json();
    renderPage();
  } catch (err) {
    console.error("Error loading movies:", err);
  }
}

function renderPage() {
  const start = (currentPage - 1) * pageSize;
  const end = start + pageSize;
  const currentMovies = movies.slice(start, end);

  const grid = document.getElementById("movie-grid");
  grid.innerHTML = "";

  currentMovies.forEach(movie => {
    const card = document.createElement("div");
    card.className = "movie-card";

    const link = document.createElement("a");
    link.href = `/movie/${movie.id}`; 

    const img = document.createElement("img");
    img.src = movie.image;
    img.alt = movie.title || movie.id;

    link.appendChild(img);
    card.appendChild(link);

    const title = document.createElement("p");
    title.textContent = movie.title || "No title";
    card.appendChild(title);

    grid.appendChild(card);
  });

  document.getElementById("page-number").textContent = `Page ${currentPage}`;
}

document.getElementById("prev").addEventListener("click", () => {
  if (currentPage > 1) {
    currentPage--;
    renderPage();
  }
});

document.getElementById("next").addEventListener("click", () => {
  const maxPage = Math.ceil(movies.length / pageSize);
  if (currentPage < maxPage) {
    currentPage++;
    renderPage();
  }
});

fetchMovies();
