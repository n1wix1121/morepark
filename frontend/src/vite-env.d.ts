/// <reference types="vite/client" />

// Декларация для CSS-модулей
declare module '*.css' {
  const content: { [className: string]: string }
  export default content
}

// Декларация для изображений
declare module '*.svg' {
  const content: string
  export default content
}

declare module '*.png' {
  const content: string
  export default content
}

declare module '*.jpg' {
  const content: string
  export default content
}