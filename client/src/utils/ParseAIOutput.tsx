import React from 'react';
import ReactMarkdown from 'react-markdown';
import type { ExtraProps } from 'react-markdown';

interface ParseMDProps {
  content: string;
}

type HeadingProps = React.ClassAttributes<HTMLHeadingElement> & React.HTMLAttributes<HTMLHeadingElement> & ExtraProps;
type ParaProps = React.ClassAttributes<HTMLParagraphElement> & React.HTMLAttributes<HTMLParagraphElement> & ExtraProps;
type ElementProps = React.ClassAttributes<HTMLElement> & React.HTMLAttributes<HTMLElement> & ExtraProps;
type ListProps = React.ClassAttributes<HTMLUListElement> & React.HTMLAttributes<HTMLUListElement> & ExtraProps;
type OListProps = React.ClassAttributes<HTMLOListElement> & React.HTMLAttributes<HTMLOListElement> & ExtraProps;
type LIProps = React.ClassAttributes<HTMLLIElement> & React.HTMLAttributes<HTMLLIElement> & ExtraProps;
type AnchorProps = React.ClassAttributes<HTMLAnchorElement> & React.HTMLAttributes<HTMLAnchorElement> & ExtraProps;
type QuoteProps = React.ClassAttributes<HTMLQuoteElement> & React.HTMLAttributes<HTMLQuoteElement> & ExtraProps;
type PreProps = React.ClassAttributes<HTMLPreElement> & React.HTMLAttributes<HTMLPreElement> & ExtraProps;

const ParseMD: React.FC<ParseMDProps> = ({ content }) => {
  const processedContent = content
    // Pull embedded score out of hadith headings: **Hadith 1 (Score: 0.6732)** → **Hadith 1**\n\nScore: 0.6732
    .replace(/\*\*(Hadith\s+\d+)\s*\(\s*Score:\s*([\d.]+)\s*\)\*\*/gi, '\n\n**$1**\n\nScore: $2')
    // Strip any remaining standalone **Score:** / **Source:** bold wrappers
    .replace(/\*\*(Source|Score):?\*\*/g, '$1:')
    // Ensure Score / Source always have a colon
    .replace(/(Source|Score)(?!:)/g, '$1:')
    // Add bold + spacing to standalone Score: and Source: lines
    .replace(/(^|\n)(Score:|Source:)/gm, '\n\n**$2**')
    // Add a hard line break after the reference marker in Source lines
    // Handles: "Hadith #584 Ali...", "Chapter 6 O Muslims...", "Sermon 3 ..."
    .replace(/((?:Hadith\s+#\d+|Chapter\s+\w+|Sermon\s+\w+))[ \t]+(?=\S)/g, '$1  \n');

  return (
    <div className="prose prose-invert max-w-none prose-p:text-on-surface prose-headings:text-on-surface prose-strong:text-on-surface prose-a:text-primary">
      <ReactMarkdown
        components={{
          h1: ({ node, ...props }: HeadingProps) => (
            <h1 className="font-[Manrope] text-2xl font-light tracking-[0.02em] text-on-surface mb-4 mt-6" {...props} />
          ),
          h2: ({ node, ...props }: HeadingProps) => (
            <h2 className="font-[Manrope] text-xl font-light tracking-[0.02em] text-on-surface mb-3 mt-5" {...props} />
          ),
          h3: ({ node, ...props }: HeadingProps) => (
            <h3 className="font-[Manrope] text-lg font-regular tracking-[0.02em] text-on-surface mb-2 mt-4" {...props} />
          ),
          p: ({ node, ...props }: ParaProps) => (
            <p className="font-[Inter] text-sm font-light leading-[1.8] text-on-surface/90 mb-3" {...props} />
          ),
          strong: ({ node, ...props }: ElementProps) => (
            <strong className="font-medium text-on-surface" {...props} />
          ),
          em: ({ node, ...props }: ElementProps) => (
            <em className="italic text-on-surface-variant" {...props} />
          ),
          ul: ({ node, ...props }: ListProps) => (
            <ul className="list-disc list-outside ml-4 space-y-1 my-3" {...props} />
          ),
          ol: ({ node, ...props }: OListProps) => (
            <ol className="list-decimal list-outside ml-4 space-y-1 my-3" {...props} />
          ),
          li: ({ node, ...props }: LIProps) => (
            <li className="font-[Inter] text-sm font-light leading-[1.7] text-on-surface/90 pl-1" {...props} />
          ),
          a: ({ node, ...props }: AnchorProps) => (
            <a className="text-primary hover:text-primary-container transition-colors duration-200 underline underline-offset-2 decoration-primary/30 hover:decoration-primary" {...props} />
          ),
          blockquote: ({ node, ...props }: QuoteProps) => (
            <blockquote className="border-l-2 border-primary-container/40 pl-4 my-4 italic text-on-surface-variant" {...props} />
          ),
          code: ({ node, ...props }: ElementProps) => (
            <code className="bg-surface-container-high rounded-[0.25rem] px-1.5 py-0.5 text-xs font-mono text-primary/80" {...props} />
          ),
          pre: ({ node, ...props }: PreProps) => (
            <pre className="bg-surface-container rounded-[0.25rem] p-4 overflow-x-auto my-4 text-sm" {...props} />
          ),
        }}
      >
        {processedContent}
      </ReactMarkdown>
    </div>
  );
};

export default ParseMD;
