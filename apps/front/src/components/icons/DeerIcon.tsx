import * as React from 'react'
import { SVGProps } from 'react'

const DeerIcon = (props: SVGProps<SVGSVGElement>) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 20"
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    strokeLinecap="round"
    strokeLinejoin="round"
    {...props}
  >
    <path stroke="none" d="M0 0h24v20H0z" fill="none" />
    <path d="M3 3c0 2 1 3 4 3c2 0 3 1 3 3" />
    <path d="M21 3c0 2 -1 3 -4 3c-2 0 -3 .333 -3 3" />
    <path d="M12 18c-1 0 -4 -3 -4 -6c0 -2 1.333 -3 4 -3s4 1 4 3c0 3 -3 6 -4 6" />
    <path d="M17 3c0 1.333 -.333 2.333 -1 3" />
    <path d="M7 3c0 1.333 .333 2.333 1 3" />
    <path d="M7 6c-2.667 .667 -4.333 1.667 -5 3" />
    <path d="M17 6c2.667 .667 4.333 1.667 5 3" />
    <path d="M8.5 10l-1.5 -1" />
    <path d="M15.5 10l1.5 -1" />
  </svg>
)
export default DeerIcon
