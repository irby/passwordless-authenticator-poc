export const MockNotificationService = {
  error: jest.fn((): string => "error"),
  info: jest.fn((): string => "info"),
  warning: jest.fn((): string => "warning"),
  success: jest.fn((): string => "success"),
  dismissibleError: jest.fn((): string => "dismiss"),
};
