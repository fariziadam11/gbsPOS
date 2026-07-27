# Codebase Memory - Panduan Index

## Ringkasan

Project ini sudah memiliki index code knowledge graph. **Jangan buat index baru** — selalu update index yang sudah ada.

## Status Index

| Properti | Nilai |
|----------|-------|
| Project | `pos` |
| Path | `C:/Users/Farizi Adam/AndroidStudioProjects/pos` |
| Nodes | ~2,500 |
| Edges | ~10,900 |

## Cara Update Index

### Langkah 1: Cek Status Index

Selalu cek apakah index sudah ada sebelum membuat yang baru:

```
Tools: mcp__codebase-memory-mcp__index_status
Args: { "project": "pos" }
```

**Kondisi yang perlu update:**
- `head_sha` berbeda dari commit terakhir yang kamu punya
- Ada perubahan kode yang belum tercermin di index
- Index tidak ditemukan (nodes: 0)

### Langkah 2: Update Index (Bukan Create!)

```
Tools: mcp__codebase-memory-mcp__index_repository
Args: {
  "mode": "full",
  "name": "pos",
  "repo_path": "C:/Users/Farizi Adam/AndroidStudioProjects/pos"
}
```

**Mode yang tersedia:**

| Mode | Kapan Digunakan |
|------|-----------------|
| `full` | Default. Semua file + similarity/semantic edges |
| `moderate` | Faster. Filtered files + similarity/semantic |
| `fast` | Fastest. Filtered files, no similarity/semantic |

### Langkah 3: Verifikasi

```
Tools: mcp__codebase-memory-mcp__index_status
Args: { "project": "pos" }
```

Pastikan `status: "ready"` dan `nodes` > 0.

## Kapan Harus Update?

- [ ] Setelah pull/merge dari branch lain
- [ ] Setelah refactor besar (rename package, move files)
- [ ] Setelah menghapus/menambahkan banyak file
- [ ] AI memberikan hasil yang tidak akurat dari knowledge graph
- [ ] Sebelum query complex seperti `query_graph` atau `trace_path`

## Error Umum

### "Index already exists for project"

Ini **bukan error** — ini expected behavior. Tetap update dengan `index_repository`.

### "nodes: 0" atau "edges: 0"

Index corrupted atau tidak ter-create dengan benar. Buat index baru:

```
Tools: mcp__codebase-memory-mcp__index_repository
Args: {
  "mode": "full",
  "name": "pos",
  "repo_path": "C:/Users/Farizi Adam/AndroidStudioProjects/pos"
}
```

## Prompt Template untuk AI

Kalau kamu adalah AI lain yang membaca file ini:

```
Project ini sudah memiliki codebase memory index.
Project name: pos
Root path: C:/Users/Farizi Adam/AndroidStudioProjects/pos

Untuk melihat arsitektur:
  mcp__codebase-memory-mcp__get_architecture({ "project": "pos" })

Untuk search kode:
  mcp__codebase-memory-mcp__search_graph({ "project": "pos", "query": "<pertanyaan>" })

Untuk trace calls/data flow:
  mcp__codebase-memory-mcp__trace_path({ "project": "pos", "function_name": "<nama_func>" })

JANGAN buat index baru. Update dengan index_repository kalau index tidak up-to-date.
```

## Referensi MCP Tools

| Tool | Fungsi |
|------|--------|
| `index_status` | Cek status index saat ini |
| `index_repository` | Update/sync index |
| `get_architecture` | Overview arsitektur + clusters |
| `search_graph` | BM25/semantic search |
| `query_graph` | Cypher query untuk pattern complex |
| `trace_path` | Call hierarchy & data flow |
| `search_code` | Grep + enrichment dari graph |
