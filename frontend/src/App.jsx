import { useEffect, useMemo, useState } from 'react'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

function useActivationCode() {
  return useMemo(() => {
    const path = window.location.pathname.replace(/\/+$/, '')
    const match = path.match(/^\/activation\/([^/]+)$/)
    return match?.[1] ?? ''
  }, [])
}

function formatStatus(status) {
  return status ? status.replaceAll('_', ' ') : 'pending activation'
}

async function requestJSON(path, options) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options?.headers ?? {}),
    },
    ...options,
  })

  const text = await response.text()
  let data = null
  try {
    data = text ? JSON.parse(text) : null
  } catch {
    data = { message: text }
  }

  if (!response.ok) {
    const message = data?.message ?? data?.data?.error ?? data?.error ?? 'Request failed'
    throw new Error(message)
  }

  return data?.data ?? null
}

export default function App() {
  const activationCode = useActivationCode()
  const [checked, setChecked] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState(null)
  const [status, setStatus] = useState(null)

  useEffect(() => {
    if (!activationCode || checked) return
    setChecked(true)
    setStatus({ message: 'Ready to activate your NETPLAY subscription.' })
  }, [activationCode, checked])

  async function handleActivate() {
    if (!activationCode) {
      setError('Activation code not found in URL.')
      return
    }

    setSubmitting(true)
    setError('')
    setStatus({ message: 'Loading...' })

    try {
      const activateResponse = await requestJSON('/api/activate', {
        method: 'POST',
        body: JSON.stringify({ activationCode }),
      })

      setResult(activateResponse)
      setStatus(activateResponse)

      const statusResponse = await requestJSON(
        `/api/subscription-status?activationCode=${encodeURIComponent(activationCode)}`,
      )
      setStatus(statusResponse)
    } catch (err) {
      setError(err.message || 'Activation failed')
      setStatus(null)
    } finally {
      setSubmitting(false)
    }
  }

  const title = result?.activationStatus === 'success' ? 'Aktivasi Sukses!' : 'Masuk ke Netplay'
  const description =
    result?.subscriptionStatus === 'active'
      ? 'Langganan Netplay Premium bulanan Anda telah aktif.'
      : 'Langganan Netplay Premium bulanan Anda akan diaktifkan pada nomor ini.'

  return (
    <main className="page-shell">
      <section className="phone-frame">
        <div className="status-bar">
          <span>9:41</span>
          <span className="status-icons">◔ ◔ ◔</span>
        </div>

        <div className="hero">
          <div className="brand-mark">
            <div className="brand-word">NETPLAY</div>
            <div className="brand-sub">ORIGINALS</div>
          </div>
        </div>

        <div className="content-card">
          <p className="eyebrow">Indico x Netplay</p>
          <h1>{title}</h1>
          <p className="lead">
            {activationCode
              ? `Kode aktivasi: ${activationCode}`
              : 'Buka halaman ini dari link aktivasi yang dikirim lewat SMS.'}
          </p>

          <div className="notice">
            <h2>Syarat & Ketentuan</h2>
            <ol>
              <li>Aktivasi hanya berlaku untuk pelanggan yang sudah terdaftar.</li>
              <li>Langganan aktif setelah tombol aktivasi ditekan dan backend memproses provider.</li>
              <li>Provider status ditampilkan setelah aktivasi berhasil.</li>
            </ol>
          </div>

          <div className="toggle-row">
            <div className="toggle-label">PEMBERITAHUAN</div>
            <p>{description}</p>
          </div>

          {error ? <div className="error-box">{error}</div> : null}
          {status?.message ? <div className="info-box">{status.message}</div> : null}

          {result?.subscriptionStatus === 'active' ? (
            <a className="cta" href="https://netplay.com" target="_blank" rel="noreferrer">
              Buka Netplay
            </a>
          ) : (
            <button className="cta" onClick={handleActivate} disabled={submitting || !activationCode}>
              {submitting ? 'Memproses...' : 'Aktivasi Sekarang'}
            </button>
          )}
        </div>
      </section>

      <section className="desktop-panel">
        <div className="desktop-card">
          <div className="desktop-logo">
            <div className="brand-word">NETPLAY</div>
            <div className="brand-sub">ORIGINALS</div>
          </div>
          <h2>Masuk ke Netplay</h2>
          <p>Aktivasi langganan premium langsung dari link yang dikirim backend.</p>
          <button className="cta" onClick={handleActivate} disabled={submitting || !activationCode}>
            {submitting ? 'Loading...' : 'Continue'}
          </button>
          <div className="desktop-meta">
            {result?.subscriptionStatus ? `Status: ${formatStatus(result.subscriptionStatus)}` : ' '}
          </div>
        </div>
      </section>
    </main>
  )
}
