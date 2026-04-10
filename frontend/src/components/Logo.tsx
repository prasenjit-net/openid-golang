interface LogoProps {
  size?: number;
  variant?: 'full' | 'icon';
  className?: string;
}

export function Logo({ size = 32, variant = 'full', className }: LogoProps) {
  return (
    <div
      className={className}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 10,
        userSelect: 'none',
        textDecoration: 'none',
      }}
    >
      {/* Shield icon with key */}
      <svg
        width={size}
        height={size}
        viewBox="0 0 40 40"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        style={{ flexShrink: 0 }}
      >
        {/* Shield shape */}
        <path
          d="M20 3L6 8.5V18C6 26.5 12.5 33.5 20 36C27.5 33.5 34 26.5 34 18V8.5L20 3Z"
          fill="url(#logoGrad)"
          stroke="rgba(255,255,255,0.15)"
          strokeWidth="0.5"
        />
        {/* Key body - circle */}
        <circle cx="18" cy="18" r="5" fill="none" stroke="white" strokeWidth="2.2" />
        {/* Key shaft */}
        <line x1="22" y1="18" x2="30" y2="18" stroke="white" strokeWidth="2.2" strokeLinecap="round" />
        {/* Key teeth */}
        <line x1="27" y1="18" x2="27" y2="21" stroke="white" strokeWidth="2" strokeLinecap="round" />
        <line x1="24" y1="18" x2="24" y2="20.5" stroke="white" strokeWidth="2" strokeLinecap="round" />
        {/* Gradient */}
        <defs>
          <linearGradient id="logoGrad" x1="6" y1="3" x2="34" y2="36" gradientUnits="userSpaceOnUse">
            <stop stopColor="#0D9488" />
            <stop offset="1" stopColor="#0F766E" />
          </linearGradient>
        </defs>
      </svg>

      {/* Text (only in 'full' variant) */}
      {variant === 'full' && (
        <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1 }}>
          <span
            style={{
              fontSize: 15,
              fontWeight: 700,
              color: 'white',
              letterSpacing: '-0.02em',
              fontFamily: 'Inter, system-ui, sans-serif',
            }}
          >
            OpenID
          </span>
          <span
            style={{
              fontSize: 10,
              fontWeight: 500,
              color: 'rgba(203,213,225,0.7)',
              letterSpacing: '0.08em',
              textTransform: 'uppercase',
              fontFamily: 'Inter, system-ui, sans-serif',
            }}
          >
            Admin Console
          </span>
        </div>
      )}
    </div>
  );
}
