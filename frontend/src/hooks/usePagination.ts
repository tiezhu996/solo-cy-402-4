import { useState } from 'react'

export function usePagination(defaultPageSize = 10) {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(defaultPageSize)
  const [total, setTotal] = useState(0)

  const onChange = (p: number, ps: number) => {
    setPage(p)
    setPageSize(ps)
  }

  return { page, pageSize, total, setTotal, onChange }
}
