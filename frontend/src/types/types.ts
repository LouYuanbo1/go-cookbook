export interface NewImageFile {
  tempID: string;
  file: File;
}

export interface ImageRequest {
  type: 'existing' | 'new' | 'deleted';
  id: number;
  tempID?: string;
  sortOrder: number;
}

export interface ImageResponse {
  id: number;
  sortOrder: number;
  imageURL: string;
}

export interface ViewIngredientCard {
  ingredientCode: string;
  name: string;
  description: string;
  image?: ImageResponse;
}

export interface ViewIngredientCardListWithCursor {
  ingredients: ViewIngredientCard[];
  cursor: number;
  hasMore: boolean;
}

export interface ViewDishCard {
	id    :   number;        
	dishCode: string ;      
	name    : string  ; 
	Image?    :ImageResponse ;
}

export interface ViewDishCardListWithCursor {
	dishes  : ViewDishCard[];
	cursor  : number;
	hasMore : boolean;
}





export interface ViewDishIngredientCard {
	id             : number        
	ingredientCode : string        
	name           : string        
	quantity       : string        
	note           : string        
	image?         : ImageResponse 
}

export interface ViewDishIngredientCardListWithCursor {
	dishIngredients : ViewDishIngredientCard[];
	cursor          : number;
  hasMore         : boolean;
}
