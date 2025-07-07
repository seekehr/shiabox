import React from 'react';
import ReactMarkdown from 'react-markdown';
import type { ExtraProps } from 'react-markdown';

interface ParseMDProps {
  content: string;
}

const ParseMD: React.FC<ParseMDProps> = ({ content }) => {
  const processedContent = content
    .replace(/\s*\*\*\s*(Source|Score)\s*:?\s*\*\*/g, '$1:')
    .replace(/(Source|Score)(?!:)/g, '$1:')
    .replace(/(Source:|Score:)/g, '\n\n**$1**');

  return (
    <div className="prose prose-invert">
      <ReactMarkdown
        components={{
          h1: ({node, ...props}: React.ClassAttributes<HTMLHeadingElement> & React.HTMLAttributes<HTMLHeadingElement> & ExtraProps) => <h1 className="text-3xl font-bold" {...props} />,
          h2: ({node, ...props}: React.ClassAttributes<HTMLHeadingElement> & React.HTMLAttributes<HTMLHeadingElement> & ExtraProps) => <h2 className="text-2xl font-bold" {...props} />,
          h3: ({node, ...props}: React.ClassAttributes<HTMLHeadingElement> & React.HTMLAttributes<HTMLHeadingElement> & ExtraProps) => <h3 className="text-xl font-bold" {...props} />,
          p: ({node, ...props}: React.ClassAttributes<HTMLParagraphElement> & React.HTMLAttributes<HTMLParagraphElement> & ExtraProps) => <p className="text-base" {...props} />,
          strong: ({node, ...props}: React.ClassAttributes<HTMLElement> & React.HTMLAttributes<HTMLElement> & ExtraProps) => <strong className="font-bold" {...props} />,
          em: ({node, ...props}: React.ClassAttributes<HTMLElement> & React.HTMLAttributes<HTMLElement> & ExtraProps) => <em className="italic" {...props} />,
          ul: ({node, ...props}: React.ClassAttributes<HTMLUListElement> & React.HTMLAttributes<HTMLUListElement> & ExtraProps) => <ul className="list-disc list-inside" {...props} />,
          ol: ({node, ...props}: React.ClassAttributes<HTMLOListElement> & React.HTMLAttributes<HTMLOListElement> & ExtraProps) => <ol className="list-decimal list-inside" {...props} />,
          li: ({node, ...props}: React.ClassAttributes<HTMLLIElement> & React.HTMLAttributes<HTMLLIElement> & ExtraProps) => <li className="my-1" {...props} />,
          a: ({node, ...props}: React.ClassAttributes<HTMLAnchorElement> & React.HTMLAttributes<HTMLAnchorElement> & ExtraProps) => <a className="text-emerald-400 hover:underline" {...props} />,
          blockquote: ({node, ...props}: React.ClassAttributes<HTMLQuoteElement> & React.HTMLAttributes<HTMLQuoteElement> & ExtraProps) => <blockquote className="border-l-4 border-gray-500 pl-4 italic" {...props} />,
          code: ({node, ...props}: React.ClassAttributes<HTMLElement> & React.HTMLAttributes<HTMLElement> & ExtraProps) => <code className="bg-gray-700 rounded px-1" {...props} />,
          pre: ({node, ...props}: React.ClassAttributes<HTMLPreElement> & React.HTMLAttributes<HTMLPreElement> & ExtraProps) => <pre className="bg-gray-700 rounded p-2 overflow-x-auto" {...props} />,
        }}
      >
        {processedContent}
      </ReactMarkdown>
    </div>
  );
};

export default ParseMD;
