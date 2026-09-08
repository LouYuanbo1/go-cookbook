// ========== 请求类型 ==========

// 更新现有图片的请求（ID + 排序顺序）
export interface UpdateImageReq {
  id: number;
  sortOrder: number;
}

// ========== 响应类型 ==========

// 基础图片响应
export interface ImageResp {
  id: number;
  sortOrder: number;
  imageURL: string;
}

// 菜品详情响应
export interface DishResp {
  dishCode: string;
  name: string;
  description: string;
  recipe: string;
  images: ImageResp[];
}

// 菜品卡片响应（列表用）
export interface DishCardResp {
  id: number;
  dishCode: string;
  name: string;
  image: ImageResp;
}

// 食材详情响应
export interface IngredientResp {
  ingredientCode: string;
  name: string;
  description: string;
  images: ImageResp[];
}

// 食材卡片响应（列表用）
export interface IngredientCardResp {
  id: number;
  ingredientCode: string;
  name: string;
  image: ImageResp;
}

// 产品详情响应
export interface ProductResp {
  productCode: string;
  ingredientCode: string;
  name: string;
  description: string;
  amount: number;
  unit: string;
  price: number;
  allergenType: string;
  images: ImageResp[];
}

// 产品卡片响应（列表用）
export interface ProductCardResp {
  id: number;
  productCode: string;
  ingredientCode: string;
  name: string;
  amount: number;
  unit: string;
  price: number;
  allergenType: string;
  image: ImageResp;
}

// 菜品食材卡片响应
export interface DishIngredientCardResp {
  id: number;
  ingredientCode: string;
  name: string;
  quantity: string;
  note: string;
  image: ImageResp;
}

// ========== 游标分页通用响应 ==========

// 通用游标响应（与后端 dto.CursorResp 对应）
export interface CursorResp<T> {
  items: T[];
  cursor: number;
  has_more: boolean;
}

// 特定类型的游标响应（便于类型推导）
export type DishCardCursorResp = CursorResp<DishCardResp>;
export type IngredientCardCursorResp = CursorResp<IngredientCardResp>;
export type DishIngredientCardCursorResp = CursorResp<DishIngredientCardResp>;
export type ProductCardCursorResp = CursorResp<ProductCardResp>;