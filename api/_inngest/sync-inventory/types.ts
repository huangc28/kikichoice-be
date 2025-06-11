export type ProductRow = {
  sku: string;
  name: string;
  ready_for_sale: boolean;
  stock_adjust_count: number;
  price: number;
  short_desc: string;
};

export type ProductWithUUID = ProductRow & {
  uuid: string;
};
