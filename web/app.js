import { fetchMovies } from './api.js'

async function renderMovies() {
  try {
    const movies = await fetchMovies()
    const container = document.getElementById('movies')
    container.innerHTML = ''

    movies.forEach(m => {
      const div = document.createElement('div')
      div.classList.add('movie-card')
      div.innerHTML = `
        <img src="${m.image}" alt="${m.title}" />
        <h3>${m.title} (${m.year})</h3>
        <a href="${m.videoUrl}" target="_blank">Watch</a>
      `
      container.appendChild(div)
    })
  } catch (e) {
    console.error(e)
  }
}

document.addEventListener('DOMContentLoaded', () => {
  renderMovies()
})

