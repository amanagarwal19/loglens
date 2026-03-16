import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts'
import { Search } from 'lucide-react'

const SERVICE_COLORS = {
  'api-gateway': '#f87171',
  'user-service': '#60a5fa',
  'order-service': '#34d399',
  'payment-service': '#fbbf24',
}

function getColor(service) {
  return SERVICE_COLORS[service] || '#a78bfa'
}

function CustomTooltip({ active, payload }) {
  if (!active || !payload || !payload.length) return null
  const d = payload[0].payload
  return (
    <div style={{ background: '#1a1a1a', border: '1px solid #2a2a2a', borderRadius: '8px', padding: '12px', fontSize: '12px', color: '#e0e0e0', maxWidth: '300px' }}>
      <p style={{ color: getColor(d.service), marginBottom: '4px' }}>{d.service}</p>
      <p style={{ fontFamily: 'monospace', marginBottom: '4px' }}>{d.message}</p>
      <p style={{ color: '#888' }}>occurrences: {d.count}</p>
    </div>
  )
}

function ClusterCard({ cluster }) {
  return (
    <div style={{ border: '1px solid #2a2a2a', borderRadius: '8px', padding: '16px', marginBottom: '12px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
        <span style={{ fontSize: '11px', padding: '2px 8px', borderRadius: '4px', color: '#000000' }}>
          {cluster.sample.level}
        </span>
        <span style={{ fontSize: '13px', color: '#000000' }}>
          {cluster.count} occurrences
        </span>
      </div>
      <div style={{ fontFamily: 'monospace', fontSize: '14px', color: '#000000', marginBottom: '8px' }}>
        {cluster.sample.message}
      </div>
      <div style={{ fontSize: '12px', color: '#666', display: 'flex', gap: '16px' }}>
        <span>service: {cluster.sample.service}</span>
        <span>first seen: {new Date(cluster.first_seen).toLocaleTimeString()}</span>
        <span>last seen: {new Date(cluster.last_seen).toLocaleTimeString()}</span>
      </div>
    </div>
  )
}

function Header() {
  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '12px',
      marginBottom: '32px',
      paddingBottom: '16px',
      borderBottom: '1px solid #e5e5e5',
      direction: 'column',
    }}>

      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '12px',
        direction: 'row',
      }}>
        <div style={{
          width: '40px',
          height: '40px',
          background: '#f87171',
          borderRadius: '10px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          direction: 'row',
          flexShrink: 0
        }}>
          <Search size={20} color="#fff" />
        </div>
        <h1 style={{ fontWeight: '500', color: '#111', margin: 20 }}>
          LogLens
        </h1>
      </div>
      <div>
        <p style={{ color: '#000000', margin: 0 }}>
          Real-time log deduplication and semantic clustering
        </p>
      </div>
    </div>
  )
}

function App() {
  const [clusters, setClusters] = useState([])

  useEffect(() => {
    const fetchClusters = () => {
      fetch('http://localhost:4000/api/clusters/all')
        .then(res => res.json())
        .then(data => setClusters(data || []))
        .catch(err => console.error('error fetching clusters:', err))
    }

    fetchClusters()
    const interval = setInterval(fetchClusters, 3000)
    return () => clearInterval(interval)
  }, [])

  // transform clusters into shape recharts expects
  const chartData = clusters
    .map(c => ({
      id: c.id,
      service: c.sample.service,
      message: c.sample.message.slice(0, 30) + '...',
      count: c.count,
    }))

  return (
    <div style={{ padding: '24px', minHeight: '100vh', color: '#000000' }}>
      <Header />
      <div style={{
        border: '1px solid #e5e5e5',
        borderRadius: '8px',
        padding: '16px',
        marginBottom: '32px'
      }}>
        <ResponsiveContainer width="100%" height={500} >
          <BarChart data={chartData} margin={{ top: 40, right: 20, bottom: 120, left: 20 }}>
            <XAxis
              dataKey="message"
              stroke="#000000"
              tick={{ fill: '#000000', fontSize: 10 }}
              angle={-35}
              textAnchor="end"
              interval={0}
              label={{
                value: 'Error type',
                position: 'insideBottom',
                offset: -110,
                fill: '#000000',
                fontSize: 18
              }}
            />
            <YAxis
              stroke="#000000"
              tick={{ fontSize: 11 }}
              label={{
                value: 'Occurrences',
                angle: -90,
                position: 'insideLeft',
                fill: '#000000',
                fontSize: 18
              }}
            />
            <Tooltip content={<CustomTooltip />} />
            <Bar dataKey="count">
              {chartData.map(entry => (
                <Cell key={entry.id} fill={getColor(entry.service)} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>

      <h2 style={{ fontSize: '16px', fontWeight: '500', margin: '32px 0 16px' }}>
        Clusters ({clusters.length})
      </h2>

      {clusters.length === 0 ? (
        <p style={{ textAlign: 'center', marginTop: '48px' }}>
          No clusters yet. Start the producer.
        </p>
      ) : (
        clusters.map(cluster => (
          <ClusterCard key={cluster.id} cluster={cluster} />
        ))
      )}
    </div>
  )
}

export default App