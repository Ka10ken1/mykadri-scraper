const API_BASE = "http://localhost:8080"


export async function fetchMovies() {
  const res = await fetch(`${API_BASE}/movies`)
  if (!res.ok) throw new Error('Failed to fetch movies')
  return res.json()
}


export async function fetchMovieImages() {
  const res = await fetch(`${API_BASE}/movie-images`)
  if (!res.ok) throw new Error('Failed to fetch movies')
  return res.json()
}

