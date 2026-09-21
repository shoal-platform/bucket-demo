'use client'

import { useEffect, useState } from 'react'

export default function Home() {
  const [items, setItems] = useState([])
  const [file, setFile] = useState(null)
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState('')

  async function loadItems() {
    const response = await fetch('/api/items', { cache: 'no-store' })
    if (!response.ok) throw new Error('Could not load bucket items')
    setItems(await response.json())
  }

  useEffect(() => { loadItems().catch(error => setMessage(error.message)) }, [])

  async function upload(event) {
    event.preventDefault()
    if (!file) return
    setBusy(true); setMessage('')
    const body = new FormData(); body.append('file', file)
    try {
      const response = await fetch('/api/upload', { method: 'POST', body })
      if (!response.ok) throw new Error(await response.text())
      setFile(null); event.target.reset(); setMessage('Uploaded successfully'); await loadItems()
    } catch (error) { setMessage(error.message) } finally { setBusy(false) }
  }

  return <main>
    <section className="hero">
      <div className="eyebrow">BUCKET // TEXT VAULT</div>
      <h1>Your words,<br /><span>stored beautifully.</span></h1>
      <p>Drop a text file into the vault, browse everything in your bucket, and download it whenever you need.</p>
    </section>
    <section className="panel">
      <form onSubmit={upload} className="upload">
        <label htmlFor="file">Upload a text file</label>
        <div className="uploadRow"><input id="file" type="file" accept=".txt,text/plain" onChange={event => setFile(event.target.files?.[0] || null)} /><button disabled={!file || busy}>{busy ? 'Uploading…' : 'Upload to bucket'}</button></div>
      </form>
      {message && <div className="message">{message}</div>}
      <div className="listHeader"><h2>Bucket items</h2><button className="refresh" onClick={() => loadItems()}>Refresh ↻</button></div>
      {items.length === 0 ? <div className="empty">Your bucket is empty. Upload your first text file above.</div> : <div className="items">{items.map(item => <div className="item" key={item.name}><div><strong>{item.name}</strong><small>{formatSize(item.size)} · {new Date(item.updated).toLocaleString()}</small></div><a href={`/api/download?name=${encodeURIComponent(item.name)}`}>Download ↓</a></div>)}</div>}
    </section>
  </main>
}

function formatSize(bytes) { return bytes < 1024 ? `${bytes} B` : `${(bytes / 1024).toFixed(1)} KB` }
