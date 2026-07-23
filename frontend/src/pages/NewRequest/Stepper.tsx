import React from 'react';

interface StepperProps {
  steps: string[];
  currentStep: number;
}

const Stepper: React.FC<StepperProps> = ({ steps, currentStep }) => (
  <ol className="stepper">
    {steps.map((label, index) => {
      const stepNumber = index + 1;
      const status =
        stepNumber < currentStep ? 'complete' : stepNumber === currentStep ? 'active' : 'upcoming';
      return (
        <li key={label} className={`stepper__step stepper__step--${status}`}>
          <span className="stepper__circle">{stepNumber < currentStep ? '✓' : stepNumber}</span>
          <span className="stepper__label">{label}</span>
          {stepNumber < steps.length && <span className="stepper__connector" aria-hidden="true" />}
        </li>
      );
    })}
  </ol>
);

export default Stepper;
