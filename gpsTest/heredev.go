package main

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/url"
)

type GeocoderResponse struct {
    Response struct {
        MetaInfo struct {
            TimeStamp string `json:"TimeStamp"`
        } `json:"MetaInfo"`
        View []struct {
            Result []struct {
                MatchLevel string `json:"MatchLevel"`
                Location   struct {
                    Address struct {
                        Label       string `json:"Label"`
                        Country     string `json:"Country"`
                        State       string `json:"State"`
                        County      string `json:"County"`
                        City        string `json:"City"`
                        District    string `json:"District"`
                        Street      string `json:"Street"`
                        HouseNumber string `json:"HouseNumber"`
                        PostalCode  string `json:"PostalCode"`
                    } `json:"Address"`
                } `json:"Location"`
            } `json:"Result"`
        } `json:"View"`
    } `json:"Response"`
}

type Position struct {
    Latitude  string `json:"latitude"`
    Longitude string `json:"longitude"`
}

type Geocoder struct {
    AppId   string `json:"app_id"`
    AppCode string `json:"app_code"`
}

func (geocoder *Geocoder) reverse(position Position) (GeocoderResponse, error) {
    endpoint, _ := url.Parse("https://reverse.geocoder.api.here.com/6.2/reversegeocode.json")
    queryParams := endpoint.Query()
    queryParams.Set("app_id", geocoder.AppId)
    queryParams.Set("app_code", geocoder.AppCode)
    queryParams.Set("mode", "retrieveAddresses")
    queryParams.Set("prox", position.Latitude+","+position.Longitude)
    endpoint.RawQuery = queryParams.Encode()
    response, err := http.Get(endpoint.String())
    if err != nil {
        return GeocoderResponse{}, err
    } else {
        data, _ := ioutil.ReadAll(response.Body)
        var geocoderResponse GeocoderResponse
        json.Unmarshal(data, &geocoderResponse)
        return geocoderResponse, nil
    }
}