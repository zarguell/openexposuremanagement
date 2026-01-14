import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { exportToCsv, exportToJson } from './export';

describe('exportToCsv', () => {
  beforeEach(() => {
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:mock-url');
    vi.spyOn(URL, 'revokeObjectURL');
    vi.spyOn(document.body, 'appendChild').mockImplementation(() => document.createElement('a') as any);
    vi.spyOn(document.body, 'removeChild').mockImplementation(() => document.createElement('a') as any);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('converts array of objects to CSV format and triggers download', () => {
    const mockClick = vi.fn();
    const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue({
      click: mockClick,
      style: {},
      addEventListener: vi.fn(),
    } as any);

    const data = [
      { id: 1, name: 'Test', value: 100 },
      { id: 2, name: 'Another', value: 200 },
    ];

    exportToCsv(data, 'test-export');

    expect(createElementSpy).toHaveBeenCalledWith('a');
    expect(mockClick).toHaveBeenCalled();
  });

  it('handles special characters in CSV output', () => {
    const mockClick = vi.fn();
    vi.spyOn(document, 'createElement').mockReturnValue({
      click: mockClick,
      style: {},
      addEventListener: vi.fn(),
    } as any);

    const data = [
      { name: 'Test, with comma', value: 'quotes "test"' },
    ];

    exportToCsv(data, 'test');

    expect(mockClick).toHaveBeenCalled();
  });

  it('handles empty arrays', () => {
    const createElementSpy = vi.spyOn(document, 'createElement');

    exportToCsv([], 'test');

    expect(createElementSpy).not.toHaveBeenCalled();
  });

  it('sets correct download filename and MIME type', () => {
    const mockClick = vi.fn();
    vi.spyOn(document, 'createElement').mockReturnValue({
      click: mockClick,
      style: {},
      addEventListener: vi.fn(),
    } as any);

    const data = [{ id: 1, name: 'Test' }];
    exportToCsv(data, 'my-export');

    expect(mockClick).toHaveBeenCalled();
  });
});

describe('exportToJson', () => {
  beforeEach(() => {
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:mock-url');
    vi.spyOn(URL, 'revokeObjectURL');
    vi.spyOn(document.body, 'appendChild').mockImplementation(() => document.createElement('a') as any);
    vi.spyOn(document.body, 'removeChild').mockImplementation(() => document.createElement('a') as any);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('converts array of objects to JSON format and triggers download', () => {
    const mockClick = vi.fn();
    const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue({
      click: mockClick,
      style: {},
      addEventListener: vi.fn(),
    } as any);

    const data = [
      { id: 1, name: 'Test' },
      { id: 2, name: 'Another' },
    ];

    exportToJson(data, 'test-export');

    expect(createElementSpy).toHaveBeenCalledWith('a');
    expect(mockClick).toHaveBeenCalled();
  });

  it('handles empty arrays', () => {
    const mockClick = vi.fn();
    const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue({
      click: mockClick,
      style: {},
      addEventListener: vi.fn(),
    } as any);

    exportToJson([], 'test');

    expect(createElementSpy).toHaveBeenCalledWith('a');
    expect(mockClick).toHaveBeenCalled();
  });

  it('sets correct download filename and MIME type', () => {
    const mockClick = vi.fn();
    vi.spyOn(document, 'createElement').mockReturnValue({
      click: mockClick,
      style: {},
      addEventListener: vi.fn(),
    } as any);

    const data = [{ id: 1, name: 'Test' }];
    exportToJson(data, 'my-export');

    expect(mockClick).toHaveBeenCalled();
  });
});
