import type { ImageRenderer, DocsPageData, BlogPageData } from '@acid-info/docusaurus-og'
import { readFileSync } from 'fs'
import { join } from 'path'
import React from 'react'
import {shapePath} from './components/ExpressiveShape/paths'

interface PagesPageData {
  route?: string
  metadata?: {
    permalink?: string
    title?: string
  }
}

const BASE_WIDTH = 1200
const BASE_HEIGHT = 630

const fonts = [
  {
    name: 'Google Sans Flex',
    data: readFileSync(
      join(__dirname, '../static/fonts/GoogleSansFlex-400.ttf'),
    ),
    weight: 400 as const,
    style: 'normal' as const,
  },
  {
    name: 'Google Sans Flex',
    data: readFileSync(
      join(__dirname, '../static/fonts/GoogleSansFlex-500.ttf'),
    ),
    weight: 500 as const,
    style: 'normal' as const,
  },
  {
    name: 'Google Sans Flex',
    data: readFileSync(
      join(__dirname, '../static/fonts/GoogleSansFlex-600.ttf'),
    ),
    weight: 600 as const,
    style: 'normal' as const,
  },
  {
    name: 'Google Sans Flex',
    data: readFileSync(
      join(__dirname, '../static/fonts/GoogleSansFlex-700.ttf'),
    ),
    weight: 700 as const,
    style: 'normal' as const,
  },
  {
    name: 'Google Sans Flex',
    data: readFileSync(
      join(__dirname, '../static/fonts/GoogleSansFlex-800.ttf'),
    ),
    weight: 800 as const,
    style: 'normal' as const,
  },
  {
    name: 'FiraCode Nerd Font Mono',
    data: readFileSync(
      join(__dirname, '../static/fonts/FiraCodeNerdFontMono-Regular.ttf'),
    ),
    weight: 400 as const,
    style: 'normal' as const,
  },
  {
    name: 'FiraCode Nerd Font Mono',
    data: readFileSync(
      join(__dirname, '../static/fonts/FiraCodeNerdFontMono-Bold.ttf'),
    ),
    weight: 700 as const,
    style: 'normal' as const,
  },
]

const logoSvg = readFileSync(
  join(__dirname, '../static/img/dux-labs-logo-dark.svg')
)

const logoDataUrl = `data:image/svg+xml;base64,${logoSvg.toString('base64')}`

const projectCategories: Record<string, string> = {
  dms: 'DMS',
  dankmaterialshell: 'DMS',
  'dms-greeter': 'DMS Greeter',
  dankgreeter: 'DMS Greeter',
  dcal: 'Dux Calendar',
  dankcalendar: 'Dux Calendar',
  duxinstall: 'Dux Install',
  dankinstall: 'Dux Install',
  dgop: 'dgop',
  dsearch: 'dsearch',
  danksearch: 'dsearch',
  packaging: 'Packaging',
  danklinux: 'Packaging',
}

function getCategoryFromPath(permalink: string): string {
  const segments = permalink.split('/').filter(Boolean)
  if (segments[0] !== 'docs') return 'Dux Labs'

  const project = /^\d+(?:\.\d+)*$/.test(segments[1]) ? segments[2] : segments[1]
  return projectCategories[project] || 'Dux Labs'
}

function titleCard(category: string, title: string): ReturnType<ImageRenderer> {
  if (title.toLowerCase().startsWith(category.toLowerCase())) {
    category = 'Dux Labs'
  }

  return [
    <div
      style={{
        display: 'flex',
        width: BASE_WIDTH,
        height: BASE_HEIGHT,
        background: 'linear-gradient(135deg, #1a1318 0%, #2d1b3d 100%)',
        position: 'relative',
        overflow: 'hidden',
        fontSmooth: 'always',
        WebkitFontSmoothing: 'antialiased',
        MozOsxFontSmoothing: 'grayscale',
      }}
    >
      <div
        style={{
          position: 'absolute',
          width: '100%',
          height: '100%',
          background: 'radial-gradient(circle at 20% 50%, rgba(128, 90, 213, 0.15) 0%, transparent 50%), radial-gradient(circle at 80% 80%, rgba(208, 188, 255, 0.12) 0%, transparent 50%)',
        }}
      />

      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          width: '100%',
          height: '100%',
          padding: '60px',
          position: 'relative',
        }}
      >
        <img
          src={logoDataUrl}
          style={{
            position: 'absolute',
            top: '50px',
            right: '50px',
            opacity: 0.6,
            width: '90px',
            height: '120px',
          }}
        />

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            flex: 1,
            justifyContent: 'center',
          }}
        >
          <div
            style={{
              fontSize: '54px',
              color: '#B794F4',
              fontFamily: 'Google Sans Flex',
              fontWeight: 500,
              marginLeft: '4px',
              marginBottom: '18px',
              letterSpacing: '0.5px',
            }}
          >
            {category}
          </div>

          <div
            style={{
              fontSize: '96px',
              color: '#ffffff',
              fontFamily: 'Google Sans Flex',
              fontWeight: 700,
              lineHeight: 1.15,
              maxWidth: '1000px',
            }}
          >
            {title}
          </div>
        </div>

        {/* <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '20px',
            borderTop: '2px solid rgba(208, 188, 255, 0.2)',
            paddingTop: '30px',
          }}
        >
          <div
            style={{
              fontSize: '36px',
              color: '#ffffff',
              fontFamily: 'Google Sans Flex',
              fontWeight: 500,
            }}
          >
            Dux
          </div>
        </div> */}
      </div>
    </div>,
    {
      width: BASE_WIDTH,
      height: BASE_HEIGHT,
      fonts,
    },
  ]
}

export const docs: ImageRenderer<DocsPageData> = (data) =>
  titleCard(getCategoryFromPath(data.metadata.permalink), data.metadata.title)

export const pages: ImageRenderer<PagesPageData> = (data) => {
  const route = data.route || data.metadata?.permalink || '/'
  const isHomePage = route === '/'
  const isPluginsPage = route === '/plugins' || route === '/plugins/'

  if (!isHomePage && !isPluginsPage) {
    const title = data.metadata?.title?.replace(/\s+\|\s+Dux$/, '') || 'Dux Labs'
    return titleCard('Dux Labs', title)
  }

  if (isPluginsPage) {
    return [
      <div
        style={{
          display: 'flex',
          width: BASE_WIDTH,
          height: BASE_HEIGHT,
          background: '#000000',
          position: 'relative',
          overflow: 'hidden',
          fontSmooth: 'always',
          WebkitFontSmoothing: 'antialiased',
          MozOsxFontSmoothing: 'grayscale',
        }}
      >
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            background: 'radial-gradient(circle at 10% 20%, rgba(128, 90, 213, 0.3) 0%, transparent 40%), radial-gradient(circle at 90% 90%, rgba(208, 188, 255, 0.05) 0%, transparent 40%), radial-gradient(circle at 50% 50%, rgba(107, 70, 193, 0.2) 0%, transparent 35%)',
          }}
        />

        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            backgroundImage: 'linear-gradient(to bottom, rgba(208, 188, 255, 0.03) 1px, transparent 1px), linear-gradient(to right, rgba(208, 188, 255, 0.03) 1px, transparent 1px)',
            backgroundSize: '50px 50px',
          }}
        />

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            width: '100%',
            height: '100%',
            padding: '60px',
            position: 'relative',
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '24px',
              marginBottom: '40px',
            }}
          >
            <img
              src={logoDataUrl}
              style={{
                width: '72px',
                height: '96px',
              }}
            />
            <div
              style={{
                fontSize: '48px',
                fontFamily: 'Google Sans Flex',
                fontWeight: 800,
                color: '#fffffff2',
                letterSpacing: '0.02em',
                textRendering: 'optimizeLegibility',
                fontSmooth: 'always',
                WebkitFontSmoothing: 'antialiased',
                MozOsxFontSmoothing: 'grayscale',
                lineHeight: 1,
              }}
            >
              Dux Labs
            </div>
          </div>

          <div
            style={{
              fontSize: '96px',
              fontWeight: 800,
              letterSpacing: '-2px',
              background: 'linear-gradient(135deg, #D0BCFF 0%, #805AD5 50%, #6B46C1 100%)',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              color: 'transparent',
              textRendering: 'optimizeLegibility',
              fontSmooth: 'always',
              WebkitFontSmoothing: 'antialiased',
              MozOsxFontSmoothing: 'grayscale',
              textAlign: 'center',
            }}
          >
            Plugins &amp; Themes
          </div>
        </div>
      </div>,
      {
        width: BASE_WIDTH,
        height: BASE_HEIGHT,
        fonts,
      },
    ]
  }

  return [
    <div
      style={{
        display: 'flex',
        width: BASE_WIDTH,
        height: BASE_HEIGHT,
        background: '#000000',
        position: 'relative',
        overflow: 'hidden',
        fontSmooth: 'always',
        WebkitFontSmoothing: 'antialiased',
        MozOsxFontSmoothing: 'grayscale',
      }}
    >
      <div
        style={{
          position: 'absolute',
          width: '100%',
          height: '100%',
          background: 'radial-gradient(circle at 10% 20%, rgba(128, 90, 213, 0.3) 0%, transparent 40%), radial-gradient(circle at 90% 90%, rgba(208, 188, 255, 0.05) 0%, transparent 40%), radial-gradient(circle at 50% 50%, rgba(107, 70, 193, 0.2) 0%, transparent 35%)',
        }}
      />

      <div
        style={{
          position: 'absolute',
          width: '100%',
          height: '100%',
          backgroundImage: 'linear-gradient(to bottom, rgba(208, 188, 255, 0.03) 1px, transparent 1px), linear-gradient(to right, rgba(208, 188, 255, 0.03) 1px, transparent 1px)',
          backgroundSize: '50px 50px',
        }}
      />

      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          width: '100%',
          height: '100%',
          padding: '60px',
          position: 'relative',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '24px',
            paddingTop: '20px',
            width: '100%',
          }}
        >
          <img
            src={logoDataUrl}
            style={{
              width: '72px',
              height: '96px',
            }}
          />
          <div
            style={{
              fontSize: '48px',
              fontFamily: 'Google Sans Flex',
              fontWeight: 800,
              color: '#fffffff2',
              letterSpacing: '0.02em',
              textRendering: 'optimizeLegibility',
              fontSmooth: 'always',
              WebkitFontSmoothing: 'antialiased',
              MozOsxFontSmoothing: 'grayscale',
              lineHeight: 1,
            }}
          >
            Dux Labs
          </div>
        </div>

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '8px',
            fontFamily: 'Google Sans Flex',
            lineHeight: 1.1,
            textAlign: 'center',
            width: '100%',
          }}
        >
          <div
            style={{
              fontSize: '96px',
              fontWeight: 800,
              letterSpacing: '-2px',
              color: '#ffffff',
              textRendering: 'optimizeLegibility',
              fontSmooth: 'always',
              WebkitFontSmoothing: 'antialiased',
              MozOsxFontSmoothing: 'grayscale',
            }}
          >
            Modern Desktop
          </div>

          <div
            style={{
              fontSize: '80px',
              fontWeight: 800,
              letterSpacing: '-2px',
              color: '#ffffff',
              textRendering: 'optimizeLegibility',
              fontSmooth: 'always',
              WebkitFontSmoothing: 'antialiased',
              MozOsxFontSmoothing: 'grayscale',
            }}
          >
            for
          </div>

          <div
            style={{
              fontSize: '96px',
              fontWeight: 800,
              letterSpacing: '-2px',
              background: 'linear-gradient(135deg, #D0BCFF 0%, #805AD5 50%, #6B46C1 100%)',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              color: 'transparent',
              textRendering: 'optimizeLegibility',
              fontSmooth: 'always',
              WebkitFontSmoothing: 'antialiased',
              MozOsxFontSmoothing: 'grayscale',
            }}
          >
            Wayland
          </div>
        </div>

        <div style={{ height: '40px' }} />
      </div>
    </div>,
    {
      width: BASE_WIDTH,
      height: BASE_HEIGHT,
      fonts,
    },
  ]
}

const heisenbergShapes = [
  {shape: 'Cookie12Sided', size: 724, left: 751, top: -304, rotation: 12, fill: 'rgba(208, 188, 255, 0.16)'},
  {shape: 'SoftBurst', size: 504, left: -191, top: 338, rotation: -8, fill: 'rgba(116, 68, 94, 0.6)'},
  {shape: 'Flower', size: 144, left: 36, top: 40, rotation: 0, fill: 'rgba(234, 223, 255, 0.08)'},
  {shape: 'Pill', size: 180, left: 960, top: 392, rotation: -40, fill: 'rgba(205, 194, 219, 0.16)'},
] as const

const heisenbergShapesSvg = `<svg xmlns="http://www.w3.org/2000/svg" width="${BASE_WIDTH}" height="${BASE_HEIGHT}">${heisenbergShapes
  .map(({shape, size, left, top, rotation, fill}) => `<path d="${shapePath(shape)}" fill="${fill}" transform="translate(${left} ${top}) rotate(${rotation} ${size / 2} ${size / 2}) scale(${size})"/>`)
  .join('')}</svg>`

const heisenbergShapesDataUrl = `data:image/svg+xml;base64,${Buffer.from(heisenbergShapesSvg).toString('base64')}`

export const heisenberg = (): ReturnType<ImageRenderer> => [
  <div style={{display: 'flex', width: BASE_WIDTH, height: BASE_HEIGHT, position: 'relative', overflow: 'hidden', backgroundColor: '#493168', color: '#eadfff', fontFamily: 'Google Sans Flex'}}>
    <img src={heisenbergShapesDataUrl} width={BASE_WIDTH} height={BASE_HEIGHT} style={{position: 'absolute', left: 0, top: 0}} />
    <div style={{display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 36, width: '100%', height: '100%', position: 'relative'}}>
      <div style={{display: 'flex', gap: 40, fontSize: 210, fontWeight: 800, lineHeight: 1, letterSpacing: '-0.04em'}}>
        <span>DMS</span>
        <span style={{color: '#d0bcff'}}>2.0</span>
      </div>
      <div style={{display: 'flex', alignItems: 'baseline', gap: 6, fontSize: 80, fontWeight: 700, lineHeight: 1}}>
        <div style={{display: 'flex', alignItems: 'center', justifyContent: 'flex-end', paddingRight: 8, position: 'relative', width: 128, height: 128, borderRadius: 16, backgroundColor: '#d0bcff', color: '#381e72'}}>
          <span style={{position: 'absolute', top: 9, left: 9, fontSize: 19, fontWeight: 500}}>2</span>
          <span style={{marginTop: 4}}>He</span>
          <span style={{position: 'absolute', bottom: 6, left: 0, right: 0, display: 'flex', justifyContent: 'center', fontSize: 14, fontWeight: 400}}>4.0026</span>
        </div>
        <span>isenberg</span>
      </div>
    </div>
  </div>,
  {width: BASE_WIDTH, height: BASE_HEIGHT, fonts},
]

// 2.1 Stormborn: the blog hero's forked lightning framing "DMS 2.1 / STORMBORN".
// Bolt geometry mirrors blog/2026-11-04-v2-1-release/StormHero.tsx (dark theme colors).
type StormPt = [number, number]

const stormBolt = (pts: StormPt[], halfW: number): string => {
  const n = pts.length - 1
  const width = (i: number) => {
    const t = i / n
    return halfW * Math.pow(1 - t, 0.85) * Math.min(1, 0.35 + t * 6)
  }
  const side = (s: 1 | -1) => pts.map(([x, y], i) => `${(x + s * width(i)).toFixed(1)} ${y.toFixed(1)}`)
  const right = side(1)
  const left = side(-1).reverse()
  return `M ${right[0]} ${right.slice(1).map((p) => `L ${p}`).join(' ')} ${left.map((p) => `L ${p}`).join(' ')} Z`
}

const stormPaths = [
  stormBolt([[60, 0], [80, 38], [50, 76], [86, 118], [56, 150], [92, 196], [70, 228], [98, 262]], 13),
  stormBolt([[50, 76], [26, 104], [38, 126], [8, 172]], 7),
  stormBolt([[86, 118], [118, 138], [106, 164], [138, 204]], 6),
  stormBolt([[56, 150], [34, 176], [44, 194], [26, 218]], 3.5),
]
const stormSparks = [[6, 178, 1.8], [140, 210, 2.2], [124, 96, 1.5], [100, 268, 2], [22, 224, 1.4], [30, 60, 1.2]]

// one bolt in its own 146x272 box, optionally mirrored
const stormBoltGroup = (mirror: boolean) =>
  `<g transform="${mirror ? 'translate(146,0) scale(-1,1)' : ''}">` +
  stormPaths.map((d) => `<path d="${d}" fill="url(#sg)"/><path d="${d}" fill="none" stroke="url(#sr)" stroke-width="0.8"/>`).join('') +
  stormSparks.map(([x, y, r]) => `<circle cx="${x}" cy="${y}" r="${r}" fill="url(#sr)" opacity="0.75"/>`).join('') +
  `</g>`

// strike box: 300px wide (aspect 146:272), placed like the hero's .strikeLeft/.strikeRight
const STORM_W = 300
const STORM_H = (STORM_W * 272) / 146
const STORM_SCALE = STORM_W / 146
const stormStrikes = [
  // top-left: origin top-left, rotate -30deg
  `translate(${-0.09 * BASE_WIDTH} ${-0.24 * BASE_HEIGHT}) rotate(-30) scale(${STORM_SCALE})`,
  // bottom-right: origin bottom-right, rotate -30deg, mirrored
  `translate(${BASE_WIDTH * 1.12} ${BASE_HEIGHT * 1.32}) rotate(-30) translate(${-STORM_W} ${-STORM_H}) scale(${STORM_SCALE})`,
]

const stormSvg = `<svg xmlns="http://www.w3.org/2000/svg" width="${BASE_WIDTH}" height="${BASE_HEIGHT}">
<defs>
  <radialGradient id="bgTop" cx="50%" cy="-20%" r="80%" fx="50%" fy="-20%"><stop offset="0%" stop-color="rgb(70,96,150)" stop-opacity="0.45"/><stop offset="60%" stop-color="rgb(70,96,150)" stop-opacity="0"/></radialGradient>
  <radialGradient id="bgBottom" cx="50%" cy="100%" r="90%"><stop offset="0%" stop-color="rgb(110,160,255)" stop-opacity="0.12"/><stop offset="70%" stop-color="rgb(110,160,255)" stop-opacity="0"/></radialGradient>
  <linearGradient id="sg" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stop-color="#ffffff" stop-opacity="0"/><stop offset="18%" stop-color="#ffffff" stop-opacity="0.98"/><stop offset="50%" stop-color="#9fd2ff" stop-opacity="0.95"/><stop offset="80%" stop-color="#5b6cff" stop-opacity="0.8"/><stop offset="100%" stop-color="#5b6cff" stop-opacity="0"/></linearGradient>
  <linearGradient id="sr" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stop-color="#f4faff" stop-opacity="0"/><stop offset="30%" stop-color="#f4faff" stop-opacity="1"/><stop offset="75%" stop-color="#9fd2ff" stop-opacity="0.85"/><stop offset="100%" stop-color="#9fd2ff" stop-opacity="0"/></linearGradient>
  <filter id="glow" x="-50%" y="-50%" width="200%" height="200%"><feGaussianBlur in="SourceGraphic" stdDeviation="9" result="b"/><feColorMatrix in="b" type="matrix" values="0 0 0 0 0.51  0 0 0 0 0.73  0 0 0 0 1  0 0 0 0.75 0" result="g"/><feMerge><feMergeNode in="g"/><feMergeNode in="SourceGraphic"/></feMerge></filter>
</defs>
<rect width="100%" height="100%" fill="#0b0f17"/>
<rect width="100%" height="100%" fill="url(#bgTop)"/>
<rect width="100%" height="100%" fill="url(#bgBottom)"/>
<g filter="url(#glow)" opacity="0.9">
  <g transform="${stormStrikes[0]}">${stormBoltGroup(false)}</g>
  <g transform="${stormStrikes[1]}">${stormBoltGroup(true)}</g>
</g>
</svg>`

const stormDataUrl = `data:image/svg+xml;base64,${Buffer.from(stormSvg).toString('base64')}`

export const stormborn = (): ReturnType<ImageRenderer> => [
  <div style={{display: 'flex', width: BASE_WIDTH, height: BASE_HEIGHT, position: 'relative', overflow: 'hidden', backgroundColor: '#0b0f17', color: '#f2f8ff', fontFamily: 'Google Sans Flex'}}>
    <img src={stormDataUrl} width={BASE_WIDTH} height={BASE_HEIGHT} style={{position: 'absolute', left: 0, top: 0}} />
    <div style={{display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 34, width: '100%', height: '100%', position: 'relative'}}>
      <div style={{display: 'flex', gap: 44, fontSize: 210, fontWeight: 800, lineHeight: 1, letterSpacing: '-0.04em', textShadow: '0 0 24px rgba(5, 7, 11, 0.85), 0 4px 12px rgba(5, 7, 11, 0.9)'}}>
        <span>DMS</span>
        <span style={{paddingRight: 10, backgroundImage: 'linear-gradient(135deg, #f2f8ff 0%, #9fd2ff 45%, #6f7dff 100%)', backgroundClip: 'text', color: 'transparent', textShadow: 'none'}}>2.1</span>
      </div>
      <div style={{display: 'flex', fontSize: 64, fontWeight: 700, lineHeight: 1, letterSpacing: '0.25em', marginRight: '-0.25em', color: '#b8dcff', textShadow: '0 0 16px rgba(5, 7, 11, 0.95), 0 2px 6px rgba(5, 7, 11, 0.95)'}}>
        STORMBORN
      </div>
    </div>
  </div>,
  {width: BASE_WIDTH, height: BASE_HEIGHT, fonts},
]

export const blog: ImageRenderer<BlogPageData> = (data) => {
  // Handle different blog page types
  let title = 'Updates'
  let isV1Release = false
  let isV12Release = false
  let isV14Release = false
  let isV20Release = false
  let isV21Release = false
  let isDankCalendar = false
  let hasStaticImage = false

  if (data.pageType === 'post' && data.data) {
    const postData = data.data as Record<string, unknown>
    const metadata = postData.metadata as Record<string, unknown> | undefined
    title = String(metadata?.title || postData.title || 'Blog')
    const id = String(postData.id || '')
    const permalink = String(metadata?.permalink || '')
    const slug = permalink.split('/').filter(Boolean).pop() || id
    isV1Release = slug === 'v1-release'
    isV12Release = slug === 'v1-2-release'
    isV14Release = slug === 'v1-4-release'
    isV20Release = slug === 'v2-0-release'
    isV21Release = slug === 'v2-1-release'
    const frontMatter = metadata?.frontMatter as Record<string, unknown> | undefined
    hasStaticImage = Boolean(frontMatter?.image)
    isDankCalendar = slug === 'dankcalendar-release'
  } else if (data.pageType === 'tag' && 'label' in data.data) {
    title = `Tag: ${String(data.data.label)}`
  } else if (data.pageType === 'archive') {
    title = 'Blog Archive'
  } else if (data.pageType === 'tags') {
    title = 'Blog Tags'
  }

  // returning undefined tells the OG plugin to keep the post's own `image` frontmatter card
  if (hasStaticImage) {
    return undefined
  }

  if (isV20Release) return heisenberg()
  if (isV21Release) return stormborn()

  // Special OG image for v1 release - matches blog hero with bats and confetti
  if (isV1Release) {
    const BatSvg = ({ color, size, opacity }: { color: string; size: number; opacity: number }) => (
      <svg
        width={size}
        height={size * 0.5}
        viewBox="0 0 100 50"
        style={{ display: 'flex' }}
      >
        <path
          d="M45 12.5 L50 17.5 L55 12.5 L60 17.5 L70 12.5 L85 7.5 L100 15 L85 25 L75 20 L65 27.5 L58 22.5 L50 32.5 L42 22.5 L35 27.5 L25 20 L15 25 L0 15 L15 7.5 L30 12.5 L40 17.5 Z"
          fill={color}
          opacity={opacity}
        />
      </svg>
    )

    return [
      <div
        style={{
          display: 'flex',
          width: BASE_WIDTH,
          height: BASE_HEIGHT,
          backgroundColor: '#0d0d0d',
          position: 'relative',
          overflow: 'hidden',
          fontSmooth: 'always',
          WebkitFontSmoothing: 'antialiased',
          MozOsxFontSmoothing: 'grayscale',
        }}
      >
        {/* Gradient overlay */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            background: 'radial-gradient(ellipse 80% 50% at 50% -20%, rgba(208, 188, 255, 0.2) 0%, transparent 60%)',
          }}
        />
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            background: 'linear-gradient(180deg, rgba(128, 90, 213, 0.06) 0%, transparent 100%)',
          }}
        />

        {/* Sparkle confetti layer */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            backgroundImage: `
              radial-gradient(circle at 10% 20%, #D0BCFF 2px, transparent 2px),
              radial-gradient(circle at 40% 70%, #805AD5 2px, transparent 2px),
              radial-gradient(circle at 70% 25%, #B794F4 1.5px, transparent 1.5px),
              radial-gradient(circle at 90% 60%, #9F7AEA 2px, transparent 2px),
              radial-gradient(circle at 25% 85%, #D0BCFF 1.5px, transparent 1.5px),
              radial-gradient(circle at 60% 45%, #805AD5 2px, transparent 2px),
              radial-gradient(circle at 85% 80%, #B794F4 1.5px, transparent 1.5px),
              radial-gradient(circle at 15% 50%, #9F7AEA 1.5px, transparent 1.5px),
              radial-gradient(circle at 22% 32%, #D0BCFF 2px, transparent 2px),
              radial-gradient(circle at 78% 55%, #805AD5 1.5px, transparent 1.5px),
              radial-gradient(circle at 52% 18%, #B794F4 2px, transparent 2px),
              radial-gradient(circle at 8% 75%, #9F7AEA 1.5px, transparent 1.5px),
              radial-gradient(circle at 88% 28%, #D0BCFF 1.5px, transparent 1.5px),
              radial-gradient(circle at 33% 58%, #805AD5 2px, transparent 2px),
              radial-gradient(circle at 67% 82%, #B794F4 1.5px, transparent 1.5px),
              radial-gradient(circle at 45% 12%, #9F7AEA 2px, transparent 2px)
            `,
            opacity: 0.5,
          }}
        />

        {/* Bat silhouettes - left side */}
        <div style={{ display: 'flex', position: 'absolute', top: '20%', left: '3%', transform: 'rotate(-15deg)' }}>
          <BatSvg color="#805AD5" size={70} opacity={0.55} />
        </div>
        <div style={{ display: 'flex', position: 'absolute', top: '44%', left: '4%', transform: 'rotate(-8deg)' }}>
          <BatSvg color="#9F7AEA" size={55} opacity={0.5} />
        </div>
        <div style={{ display: 'flex', position: 'absolute', top: '68%', left: '5%', transform: 'rotate(-20deg)' }}>
          <BatSvg color="#B794F4" size={50} opacity={0.45} />
        </div>

        {/* Bat silhouettes - right side */}
        <div style={{ display: 'flex', position: 'absolute', top: '22%', left: '88%', transform: 'rotate(12deg)' }}>
          <BatSvg color="#B794F4" size={60} opacity={0.5} />
        </div>
        <div style={{ display: 'flex', position: 'absolute', top: '46%', left: '86%', transform: 'rotate(18deg)' }}>
          <BatSvg color="#805AD5" size={75} opacity={0.55} />
        </div>
        <div style={{ display: 'flex', position: 'absolute', top: '70%', left: '89%', transform: 'rotate(8deg)' }}>
          <BatSvg color="#9F7AEA" size={55} opacity={0.5} />
        </div>

        {/* Main content */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            width: '100%',
            height: '100%',
            position: 'relative',
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'baseline',
              gap: '24px',
              fontFamily: 'Google Sans Flex',
            }}
          >
            <div
              style={{
                fontSize: '180px',
                fontWeight: 800,
                color: '#ffffff',
                letterSpacing: '-4px',
              }}
            >
              DMS
            </div>
            <div
              style={{
                fontSize: '180px',
                fontWeight: 800,
                letterSpacing: '-4px',
                background: 'linear-gradient(135deg, #D0BCFF 0%, #805AD5 100%)',
                backgroundClip: 'text',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                color: 'transparent',
              }}
            >
              1.0
            </div>
          </div>

          <div
            style={{
              fontSize: '32px',
              fontWeight: 600,
              color: 'rgba(208, 188, 255, 0.4)',
              fontFamily: 'Google Sans Flex',
              marginTop: '16px',
              letterSpacing: '0.25em',
              textTransform: 'uppercase',
            }}
          >
            The Dark Knight
          </div>
        </div>
      </div>,
      {
        width: BASE_WIDTH,
        height: BASE_HEIGHT,
        fonts,
      },
    ]
  }

  // Special OG image for v1.4 release - Saffron Bloom
  if (isV14Release) {
    return [
      <div
        style={{
          display: 'flex',
          width: BASE_WIDTH,
          height: BASE_HEIGHT,
          background: '#000000',
          position: 'relative',
          overflow: 'hidden',
          fontSmooth: 'always',
          WebkitFontSmoothing: 'antialiased',
          MozOsxFontSmoothing: 'grayscale',
        }}
      >
        {/* Radial gradient overlays - saffron/crimson tones */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            background: 'radial-gradient(circle at 10% 20%, rgba(139, 30, 30, 0.3) 0%, transparent 40%), radial-gradient(circle at 90% 90%, rgba(232, 168, 50, 0.08) 0%, transparent 40%), radial-gradient(circle at 50% 50%, rgba(107, 21, 21, 0.2) 0%, transparent 35%)',
          }}
        />

        {/* Grid pattern */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            backgroundImage: 'linear-gradient(to bottom, rgba(232, 168, 50, 0.03) 1px, transparent 1px), linear-gradient(to right, rgba(232, 168, 50, 0.03) 1px, transparent 1px)',
            backgroundSize: '50px 50px',
          }}
        />

        {/* Confetti sparkles - saffron color palette */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            backgroundImage: `
              radial-gradient(circle at 8% 15%, #E8A832 2.5px, transparent 2.5px),
              radial-gradient(circle at 15% 75%, #8B1E1E 2px, transparent 2px),
              radial-gradient(circle at 25% 35%, #C65D3B 2px, transparent 2px),
              radial-gradient(circle at 35% 85%, #D4912A 2.5px, transparent 2.5px),
              radial-gradient(circle at 45% 20%, #6B1515 2px, transparent 2px),
              radial-gradient(circle at 55% 70%, #8B1E1E 2.5px, transparent 2.5px),
              radial-gradient(circle at 65% 25%, #C65D3B 2px, transparent 2px),
              radial-gradient(circle at 75% 80%, #E8A832 2px, transparent 2px),
              radial-gradient(circle at 85% 40%, #6B1515 2.5px, transparent 2.5px),
              radial-gradient(circle at 92% 65%, #8B1E1E 2px, transparent 2px),
              radial-gradient(circle at 12% 45%, #D4912A 2px, transparent 2px),
              radial-gradient(circle at 88% 18%, #C65D3B 2px, transparent 2px),
              radial-gradient(circle at 42% 90%, #E8A832 2px, transparent 2px),
              radial-gradient(circle at 72% 12%, #8B1E1E 2.5px, transparent 2.5px),
              radial-gradient(circle at 5% 55%, #6B1515 2.5px, transparent 2.5px),
              radial-gradient(circle at 95% 85%, #D4912A 2.5px, transparent 2.5px)
            `,
            opacity: 0.6,
          }}
        />

        {/* Main content */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            width: '100%',
            height: '100%',
            position: 'relative',
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'baseline',
              gap: '24px',
              fontFamily: 'Google Sans Flex',
            }}
          >
            <div
              style={{
                fontSize: '180px',
                fontWeight: 800,
                color: '#ffffff',
                letterSpacing: '-4px',
              }}
            >
              DMS
            </div>
            <div
              style={{
                fontSize: '180px',
                fontWeight: 800,
                letterSpacing: '-4px',
                background: 'linear-gradient(135deg, #E8A832 0%, #C65D3B 50%, #8B1E1E 100%)',
                backgroundClip: 'text',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                color: 'transparent',
              }}
            >
              1.4
            </div>
          </div>

          <div
            style={{
              fontSize: '32px',
              fontWeight: 600,
              color: 'rgba(198, 93, 59, 0.4)',
              fontFamily: 'Google Sans Flex',
              marginTop: '16px',
              letterSpacing: '0.25em',
              textTransform: 'uppercase',
            }}
          >
            Saffron Bloom
          </div>
        </div>
      </div>,
      {
        width: BASE_WIDTH,
        height: BASE_HEIGHT,
        fonts,
      },
    ]
  }

  // Special OG image for v1.2 release - Spicy Miso
  if (isV12Release) {
    return [
      <div
        style={{
          display: 'flex',
          width: BASE_WIDTH,
          height: BASE_HEIGHT,
          background: '#000000',
          position: 'relative',
          overflow: 'hidden',
          fontSmooth: 'always',
          WebkitFontSmoothing: 'antialiased',
          MozOsxFontSmoothing: 'grayscale',
        }}
      >
        {/* Radial gradient overlays - warm red/orange tones */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            background: 'radial-gradient(circle at 10% 20%, rgba(239, 68, 68, 0.3) 0%, transparent 40%), radial-gradient(circle at 90% 90%, rgba(251, 146, 60, 0.08) 0%, transparent 40%), radial-gradient(circle at 50% 50%, rgba(220, 38, 38, 0.2) 0%, transparent 35%)',
          }}
        />

        {/* Grid pattern */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            backgroundImage: 'linear-gradient(to bottom, rgba(251, 146, 60, 0.03) 1px, transparent 1px), linear-gradient(to right, rgba(251, 146, 60, 0.03) 1px, transparent 1px)',
            backgroundSize: '50px 50px',
          }}
        />

        {/* Confetti sparkles - red/orange tones */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            backgroundImage: `
              radial-gradient(circle at 8% 15%, #ef4444 2.5px, transparent 2.5px),
              radial-gradient(circle at 15% 75%, #dc2626 2px, transparent 2px),
              radial-gradient(circle at 25% 35%, #f97316 2px, transparent 2px),
              radial-gradient(circle at 35% 85%, #fb923c 2.5px, transparent 2.5px),
              radial-gradient(circle at 45% 20%, #ef4444 2px, transparent 2px),
              radial-gradient(circle at 55% 70%, #dc2626 2.5px, transparent 2.5px),
              radial-gradient(circle at 65% 25%, #f97316 2px, transparent 2px),
              radial-gradient(circle at 75% 80%, #fb923c 2px, transparent 2px),
              radial-gradient(circle at 85% 40%, #ef4444 2.5px, transparent 2.5px),
              radial-gradient(circle at 92% 65%, #dc2626 2px, transparent 2px),
              radial-gradient(circle at 12% 45%, #f97316 2px, transparent 2px),
              radial-gradient(circle at 88% 18%, #fb923c 2px, transparent 2px),
              radial-gradient(circle at 42% 90%, #ef4444 2px, transparent 2px),
              radial-gradient(circle at 72% 12%, #dc2626 2.5px, transparent 2.5px),
              radial-gradient(circle at 5% 55%, #f97316 2.5px, transparent 2.5px),
              radial-gradient(circle at 95% 85%, #fb923c 2.5px, transparent 2.5px)
            `,
            opacity: 0.6,
          }}
        />

        {/* Main content */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            width: '100%',
            height: '100%',
            position: 'relative',
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'baseline',
              gap: '24px',
              fontFamily: 'Google Sans Flex',
            }}
          >
            <div
              style={{
                fontSize: '180px',
                fontWeight: 800,
                color: '#ffffff',
                letterSpacing: '-4px',
              }}
            >
              DMS
            </div>
            <div
              style={{
                fontSize: '180px',
                fontWeight: 800,
                letterSpacing: '-4px',
                background: 'linear-gradient(135deg, #fca5a5 0%, #ef4444 50%, #dc2626 100%)',
                backgroundClip: 'text',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                color: 'transparent',
              }}
            >
              1.2
            </div>
          </div>

          <div
            style={{
              fontSize: '32px',
              fontWeight: 600,
              color: 'rgba(252, 165, 165, 0.4)',
              fontFamily: 'Google Sans Flex',
              marginTop: '16px',
              letterSpacing: '0.25em',
              textTransform: 'uppercase',
            }}
          >
            Spicy Miso
          </div>
        </div>
      </div>,
      {
        width: BASE_WIDTH,
        height: BASE_HEIGHT,
        fonts,
      },
    ]
  }

  // Special OG image for the DankCalendar launch - matches the blog hero
  if (isDankCalendar) {
    const CalTile = ({ day, size, left, top, rot, opacity }: { day: string; size: number; left: string; top: string; rot: number; opacity: number }) => (
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'absolute',
          left,
          top,
          width: size,
          height: size,
          borderRadius: size * 0.18,
          border: '2px solid rgba(128, 90, 213, 0.45)',
          background: 'rgba(128, 90, 213, 0.12)',
          color: '#B794F4',
          fontFamily: 'Google Sans Flex',
          fontWeight: 700,
          fontSize: size * 0.42,
          transform: `rotate(${rot}deg)`,
          opacity,
        }}
      >
        <div
          style={{
            display: 'flex',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            height: size * 0.22,
            background: 'rgba(128, 90, 213, 0.55)',
            borderRadius: `${size * 0.18}px ${size * 0.18}px 0 0`,
          }}
        />
        {day}
      </div>
    )

    return [
      <div
        style={{
          display: 'flex',
          width: BASE_WIDTH,
          height: BASE_HEIGHT,
          backgroundColor: '#0d0d0d',
          position: 'relative',
          overflow: 'hidden',
          fontSmooth: 'always',
          WebkitFontSmoothing: 'antialiased',
          MozOsxFontSmoothing: 'grayscale',
        }}
      >
        {/* Gradient overlays */}
        <div
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            background: 'radial-gradient(ellipse 80% 50% at 50% -20%, rgba(128, 90, 213, 0.25) 0%, transparent 60%), radial-gradient(circle at 85% 85%, rgba(99, 179, 237, 0.1) 0%, transparent 45%)',
          }}
        />

        {/* Floating calendar tiles - kept to the edges */}
        <CalTile day="15" size={88} left="6%" top="16%" rot={-12} opacity={0.85} />
        <CalTile day="3" size={70} left="11%" top="52%" rot={8} opacity={0.7} />
        <CalTile day="28" size={64} left="4%" top="78%" rot={15} opacity={0.6} />
        <CalTile day="21" size={80} left="84%" top="15%" rot={10} opacity={0.8} />
        <CalTile day="6" size={66} left="90%" top="50%" rot={-14} opacity={0.65} />
        <CalTile day="30" size={76} left="82%" top="78%" rot={12} opacity={0.75} />

        {/* Main content */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            width: '100%',
            height: '100%',
            position: 'relative',
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'baseline',
              fontFamily: 'Google Sans Flex',
            }}
          >
            <div
              style={{
                fontSize: '150px',
                fontWeight: 800,
                color: '#ffffff',
                letterSpacing: '-4px',
              }}
            >
              Dank
            </div>
            <div
              style={{
                fontSize: '150px',
                fontWeight: 800,
                letterSpacing: '-4px',
                background: 'linear-gradient(135deg, #D0BCFF 0%, #805AD5 100%)',
                backgroundClip: 'text',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                color: 'transparent',
              }}
            >
              Calendar
            </div>
          </div>

          <div
            style={{
              fontSize: '32px',
              fontWeight: 600,
              color: 'rgba(183, 148, 244, 0.5)',
              fontFamily: 'Google Sans Flex',
              marginTop: '16px',
              letterSpacing: '0.25em',
              textTransform: 'uppercase',
            }}
          >
            0.1 · First Release
          </div>
        </div>
      </div>,
      {
        width: BASE_WIDTH,
        height: BASE_HEIGHT,
        fonts,
      },
    ]
  }

  return titleCard('Updates', title)
}
